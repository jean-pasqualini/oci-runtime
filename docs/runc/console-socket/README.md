Top, tu veux faire un --console-socket “à la runc” dans ton runc maison. L’idée générale :
•	Si spec.process.terminal = true, tu crées un PTY (master/slave).
•	Tu branches le slave comme stdin/stdout/stderr + TTY contrôlant du process init dans le conteneur.
•	Tu envoies le FD du master via un Unix Domain Socket (SCM_RIGHTS) au chemin donné par --console-socket.
Le receveur (shim, ou ton outil) pourra alors lire/écrire la console.

Ci-dessous, le déroulé concret + un squelette Go prêt à adapter.

⸻

Ce que fait --console-socket (comportement attendu)
1.	Préconditions

	•	Flag fourni uniquement si Terminal=true dans la spec OCI.
	•	Le socket UNIX pointé par --console-socket doit déjà écouter (c’est le côté “shim”).

	2.	Création du PTY

	•	openpty() → obtient (master, slave).
	•	Dans le process enfant (init du conteneur) :
	•	setsid() pour créer une nouvelle session.
	•	ioctl(TIOCSCTTY, slave) pour en faire le tty contrôlant.
	•	dup2(slave → 0,1,2) pour stdio.
	•	fermer les FDs surnuméraires.
	•	Monter/présenter le tty comme /dev/console dans le conteneur (selon ta stack / devpts).

	3.	Transfert du master à l’extérieur

	•	Dans le parent (sur l’hôte), ouvrir le socket --console-socket.
	•	Envoyer master via SCM_RIGHTS (sendmsg).
Optionnel : petit octet/ACK pour synchroniser.
•	Fermer ta copie locale du master si tu ne t’en sers pas.

	4.	Resize (optionnel mais utile)

	•	Fournir un petit endpoint (ou une commande) pour faire un ioctl(TIOCSWINSZ) côté conteneur ou côté master, selon ton design.

⸻

Squelette Go (minimal, sans gestion d’erreurs verbeuse)

Dépendances utiles :
github.com/creack/pty (ou appelle unix.Openpty) et golang.org/x/sys/unix.

```go
package main

import (
"context"
"log"
"log/slog"
"net"
"os"
"os/exec"
"syscall"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

func sendFDOverUnixSocket(sockPath string, fd int) error {
conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: sockPath, Net: "unix"})
if err != nil {
return err
}
defer conn.Close()

	// On a besoin du RawConn pour sendmsg + SCM_RIGHTS
	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	var opErr error
	err = raw.Control(func(s uintptr) {
		// On envoie 1 octet de payload + ancillaire contenant le FD
		oob := unix.UnixRights(fd)
		_, _, e := unix.Sendmsg(int(s), []byte{0x00}, oob, nil, 0)
		if e != nil {
			opErr = e
			return
		}
	})
	if err != nil {
		return err
	}
	return opErr
}

func runWithConsoleSocket(specPath, consoleSock string, logger *slog.Logger) error {
// 1) Charger ta spec OCI (omise ici) et vérifier Process.Terminal == true
terminal := true // <- remplace par lecture spec
if !terminal && consoleSock != "" {
return fmt.Errorf("--console-socket fourni mais Terminal=false")
}

	// 2) Créer le PTY
	master, slave, err := pty.Open()
	if err != nil {
		return err
	}
	defer master.Close() // on fermera plus tard après envoi du FD

	// 3) Préparer le process enfant (init du conteneur)
	// Remplace /init par ton binaire init “inside container”
	cmd := exec.Command("/init")
	// On va attacher le slave aux 3 stdio du child
	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave

	// IMPORTANT: setsid + TIOCSCTTY doivent s'appliquer au child.
	// On peut utiliser SysProcAttr pour setsid, puis un pre-start hook pour TIOCSCTTY.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true, // nouvelle session, nécessaire pour controlling TTY
		// Plus tard: Cloneflags pour namespaces, Credential, Chroot, etc.
	}

	// 4) Lancement du child
	if err := cmd.Start(); err != nil {
		slave.Close()
		return err
	}
	childPid := cmd.Process.Pid

	// Côté parent, on peut fermer le slave (le child l'a)
	slave.Close()

	// 5) Faire de slave le controlling TTY -> TIOCSCTTY
	//    Ici, on est dans le parent; pour être strict, fais-le côté child via un code wrapper/init
	//    ou un PR_SET_PDEATHSIG+petit binaire qui fait ioctl(TIOCSCTTY).
	//    Pour un exemple rapide, on peut faire l’ioctl ici si on a encore le fd slave dans le child,
	//    mais le plus propre est de le faire *dans* le child avant exec.

	// 6) Envoyer le FD master via --console-socket
	if consoleSock != "" {
		if err := sendFDOverUnixSocket(consoleSock, int(master.Fd())); err != nil {
			logger.Error("send master fd failed", "err", err)
			// À toi de décider: continuer quand même ou abort
		}
		// Si c'est juste un transfert, tu peux fermer ta copie locale:
		// master.Close()
	}

	logger.Info("container init started", "pid", childPid)

	// 7) (Optionnel) Attendre l'init ou retourner immédiatement
	return nil
}

func main() {
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
if err := runWithConsoleSocket("/path/to/config.json", "/run/myshim/console.sock", logger); err != nil {
log.Fatal(err)
}
}
```

Notes importantes
•	TIOCSCTTY : le controlling terminal doit être défini dans le process enfant (celui qui a Setsid). Le plus propre : un petit wrapper (ton “init inside”) qui fait :
•	setsid()
•	ioctl(slave, TIOCSCTTY)
•	dup2(slave → 0,1,2)
•	puis exec de l’entrée de ta spec.
•	Namespaces & chroot : dans un runc “vrai”, tu configures tout (CLONE_NEW*, mount namespace, pivot_root, etc.) avant d’exec ton init. Ici j’ai laissé ça de côté pour se focaliser sur la console.
•	/dev/console : assure-toi que ton devpts est monté dans le mount namespace du conteneur et que /dev/console pointe vers le tty attendu (souvent un bind vers le slave pty).
•	Protocole socket : minimaliste (juste un sendmsg du FD). Tu peux ajouter un ACK (le receveur envoie un octet “OK”) pour confirmer la réception avant de fermer master.
•	Resize : expose une commande/endpoint pour faire :

ws := &unix.Winsize{Col: cols, Row: rows}
_ = unix.IoctlSetWinsize(int(master.Fd()), unix.TIOCSWINSZ, ws)


	•	Erreurs & permissions : le socket doit être accessible par le user qui lance ton runc maison; pense à l’umask/owner.

⸻

Checklist côté “shim” (récepteur du FD)
•	Créer/écouter sur --console-socket (UNIX datagram ou stream).
•	À la réception : récupérer le FD master via recvmsg + SCM_RIGHTS.
•	Lire/écrire dessus pour faire attach, logs, resize, etc.
•	Fermer proprement quand le conteneur meurt.

⸻

Si tu me dis comment tu organises l’init (wrapper côté conteneur ou tout fait côté parent), je peux te donner un flow exact avec setsid/TIOCSCTTY au bon endroit + un mini “init” de 20 lignes ✨