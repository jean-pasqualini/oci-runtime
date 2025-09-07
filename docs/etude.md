Très bonne question 👌

Quand tu fais un crun run (ou runc run) avec --tty/mode attach, il y a quelques points clés pour que ça marche comme dans un vrai terminal :

⸻

🔹 Ce que fait crun en mode attach
1.	Créer un pseudo-terminal (pty)
•	côté “esclave” → attaché au process du conteneur (ex: /dev/pts/N devient son stdin/stdout/stderr).
•	côté “maître” → gardé par le runtime, et relié au terminal de l’utilisateur.
2.	Brancher les flux
•	L’utilisateur tape → ça va au pty maître → redirigé au process dans le conteneur.
•	Le process écrit → ça sort via le pty maître → affiché dans ton terminal.
3.	Configurer les signaux
•	Ex : Ctrl-C (SIGINT), Ctrl-Z… sont forwardés au process du conteneur.
•	Redimensionnement (SIGWINCH) → la taille du terminal est synchronisée (si tu fais resize ou tmux split, ça suit).
4.	Mode terminal brut (raw mode)
•	Quand tu attaches, le runtime met le terminal hôte en mode raw (désactive l’écho, line buffering, etc.), sinon tu aurais un double écho.

⸻

🔹 En Go : équivalent minimal

Si tu veux imiter ce comportement dans ton RunHandler :
1.	Allouer un pty

import "github.com/creack/pty"

cmd := exec.CommandContext(ctx, "/bin/bash")
ptmx, err := pty.Start(cmd) // lance le process avec un pty
if err != nil { return err }
defer func() { _ = ptmx.Close() }()


	2.	Relay I/O avec ton terminal

go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
_, _ = io.Copy(os.Stdout, ptmx)


	3.	Forwarder les signaux (Ctrl-C, resize)

c := make(chan os.Signal, 1)
signal.Notify(c, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)
go func() {
for sig := range c {
switch sig {
case syscall.SIGWINCH:
_ = pty.InheritSize(os.Stdin, ptmx) // resize
default:
_ = cmd.Process.Signal(sig) // forward
}
}
}()
// init resize au démarrage
_ = pty.InheritSize(os.Stdin, ptmx)



⸻

🔹 Résumé : ce qu’il faut pour “attach” façon crun
•	Un pty (pas juste des pipes).
•	Redirection bidirectionnelle stdin/stdout/stderr ↔ pty ↔ conteneur.
•	Mode raw sur le terminal hôte (golang.org/x/term.MakeRaw).
•	Signal forwarding (SIGWINCH pour resize, SIGINT pour ^C, etc.).
•	Cleanup quand le process se termine (remettre le terminal en mode normal).

⸻

👉 Veux-tu que je te prépare un handler RunCmd version attach déjà câblé avec github.com/creack/pty, histoire que tu aies un prototype qui se comporte comme crun run --tty ?


----

master, slave, err := unix.Openpty(nil, nil) // open a new pty
cmd := exec.Command("/bin/bash")
cmd.Stdin = slave
cmd.Stdout = slave
cmd.Stderr = slave

----

Awesome—here’s a minimal, idiomatic way to make a Run handler that launches a bash process in new Linux namespaces (no heavy “runtime” yet), using your generic handler + decorators style.

I’ll show two variants:
•	A. Rooted (simplest): needs root (or CAP_SYS_ADMIN).
•	B. Unprivileged with user namespaces (works as normal user on kernels that allow it).

⸻

A) Rooted: new PID + mount + UTS + IPC namespaces

// internal/app/run.go
package app

import (
"context"
"os"
"os/exec"
"syscall"
"time"
)

type RunCmd struct {
// Optional: override the shell command; empty -> interactive bash
Command  []string
Hostname string // optional: set container hostname
Timeout  time.Duration
}

func NewRun() HandlerFunc[RunCmd] {
return func(ctx context.Context, cmd RunCmd) error {
// Build the shell line: mount /proc inside the new mntns, set hostname, then exec bash (or your command)
script := ""
if cmd.Hostname != "" {
script += "hostname " + cmd.Hostname + " && "
}
// mount /proc so `ps` etc. work inside the new PID namespace
script += "mount -t proc proc /proc && "

		if len(cmd.Command) > 0 {
			// run your command
			script += "exec " + joinArgs(cmd.Command)
		} else {
			// default interactive shell
			script += "exec bash"
		}

		c := exec.CommandContext(ctx, "/bin/bash", "-lc", script)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.SysProcAttr = &syscall.SysProcAttr{
			// new UTS (hostname), new PID ns, new mount ns, new IPC
			Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
			Unshareflags: syscall.CLONE_NEWNS, // ensure a private mount ns before we mount /proc
		}
		return c.Run()
	}
}

// tiny helper: turn []string into a single shell-escaped command
func joinArgs(args []string) string {
// keep it simple for a trusted input; for untrusted, use proper escaping.
switch len(args) {
case 0:
return ""
case 1:
return args[0]
default:
// join with spaces; caller provides already-split tokens
s := args[0]
for _, a := range args[1:] {
s += " " + a
}
return s
}
}

Usage (with your middleware chain):

runH := app.Chain(
app.NewRun(),
app.WithTimeout[app.RunCmd](30*time.Second),
// app.WithLogging[app.RunCmd](logger, "run"),
)

_ = runH(ctx, app.RunCmd{
Hostname: "ns-demo",
// Command: []string{"bash", "-l"}, // or custom
})

Notes
•	Requires root (or appropriate caps) for NEWNS/NEWPID/NEWUTS/NEWIPC.
•	We mount /proc inside the new mntns so tools work.
•	CLONE_NEWPID means your bash becomes PID 1 in its namespace; if you run daemons, consider reaping zombies.

⸻

B) Unprivileged: add user namespace + ID mappings

If you want this to work as a regular user:

c.SysProcAttr = &syscall.SysProcAttr{
Cloneflags:   syscall.CLONE_NEWUSER | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
Unshareflags: syscall.CLONE_NEWNS,
// Map container uid/gid 0 -> your real uid/gid on the host
UidMappings: []syscall.SysProcIDMap{{
ContainerID: 0, HostID: os.Getuid(), Size: 1,
}},
GidMappingsEnableSetgroups: false, // required before setting GidMappings on many kernels
GidMappings: []syscall.SysProcIDMap{{
ContainerID: 0, HostID: os.Getgid(), Size: 1,
}},
}

Notes
•	Kernel must allow unprivileged user namespaces (it’s common, but can be disabled).
•	Inside the userns you have a kind of “fake root” (uid 0 mapped to your real uid). Some mounts may still require extra allowances.

⸻

Signal handling & cleanup (optional starters)
•	Forward signals from the parent to child if you want Ctrl-C to work for the bash: by default, attached TTY will do, but for programmatic signals use cmd.Process.Signal(...).
•	If you want to detach, consider cmd.SysProcAttr.Setsid = true and manage a pty.

⸻

Testing this handler
•	Unit-test: pass a fake exec via a seam (e.g., wrap exec.CommandContext behind a small interface so you can assert SysProcAttr flags).
•	Integration-test (requires Linux): run it under root (or with userns) and check uname -n, readlink /proc/1/exe, etc.

If you want, I can refactor this into a slightly more testable shape (injecting a CommandBuilder interface), or add a variant that chroots/pivot_root to a given rootfs/ bundle before exec.


------

Voici une version “à la runc” (sans creack/pty), qui utilise directement les syscalls Linux via golang.org/x/sys/unix et met le terminal en raw mode avec golang.org/x/term.

Points clés :
•	unix.Openpty pour allouer le PTY (master/slave)
•	unix.IoctlGetWinsize / unix.IoctlSetWinsize pour la taille
•	term.MakeRaw / term.Restore pour le mode raw du terminal hôte
•	redirection I/O bidirectionnelle + gestion SIGWINCH (resize)
•	Linux-only (build tag)

//go:build linux

package app

import (
"context"
"fmt"
"io"
"os"
"os/exec"
"os/signal"
"syscall"
"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// RunAttach lance argv[0] (ex: "/bin/bash") attaché à un PTY, façon runc/crun.
func RunAttach(ctx context.Context, argv []string) error {
if len(argv) == 0 {
return fmt.Errorf("argv empty (e.g. ['/bin/bash'])")
}

	// 1) Créer le PTY
	mFD, sFD, err := unix.Openpty(nil, nil)
	if err != nil {
		return fmt.Errorf("openpty: %w", err)
	}
	defer unix.Close(mFD)
	defer unix.Close(sFD)

	master := os.NewFile(uintptr(mFD), "/dev/ptmx")
	slave  := os.NewFile(uintptr(sFD), "/dev/pts/slave")
	defer master.Close()
	defer slave.Close()

	// 2) Mettre le terminal hôte en raw + restaurer à la fin
	var oldState *term.State
	if term.IsTerminal(int(os.Stdin.Fd())) {
		if st, err := term.MakeRaw(int(os.Stdin.Fd())); err == nil {
			oldState = st
			defer term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}

	// 3) Synchroniser la taille du PTY avec le terminal courant
	if err := inheritSize(int(os.Stdin.Fd()), int(master.Fd())); err != nil {
		// pas bloquant, on continue
	}

	// 4) Préparer la commande
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	// Attacher le slave au process enfant
	cmd.Stdin  = slave
	cmd.Stdout = slave
	cmd.Stderr = slave

	// Optionnel: si tu veux des namespaces, ajoute SysProcAttr ici.
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
	// 	Unshareflags: syscall.CLONE_NEWNS,
	// }

	// Important: assure que le FD "slave" devient le ctty du process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,                       // nouvelle session
		Setctty: true,                       // définir le ctty
		Ctty:    int(slave.Fd()),            // ctty = pty slave
	}

	// 5) Gestion du resize (SIGWINCH)
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)
	go func() {
		for range winch {
			_ = inheritSize(int(os.Stdin.Fd()), int(master.Fd()))
		}
	}()
	// resize initial
	_ = inheritSize(int(os.Stdin.Fd()), int(master.Fd()))

	// 6) Démarrer le process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	// Ferme le côté slave dans le parent, l’enfant garde son côté
	_ = slave.Close()

	// 7) Relayer l’I/O
	// stdin → master
	go func() { _, _ = io.Copy(master, os.Stdin) }()
	// master → stdout
	copyDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(os.Stdout, master)
		close(copyDone)
	}()

	// 8) Attendre la fin ou le contexte
	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	select {
	case err := <-waitDone:
		// Process terminé : attendre que la copie stdout finisse
		select {
		case <-copyDone:
		case <-time.After(200 * time.Millisecond):
		}
		return err
	case <-ctx.Done():
		// Contexte annulé : tenter de tuer proprement
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case err := <-waitDone:
			return err
		case <-time.After(2 * time.Second):
			_ = cmd.Process.Kill()
			return ctx.Err()
		}
	}
}

func inheritSize(fromFD, toMasterFD int) error {
ws, err := unix.IoctlGetWinsize(fromFD, unix.TIOCGWINSZ)
if err != nil {
return err
}
return unix.IoctlSetWinsize(toMasterFD, unix.TIOCSWINSZ, ws)
}

Notes rapides
•	TTY/Raw : on met le TTY hôte en mode brut pour un vrai comportement interactif (comme runc --tty).
•	PTY : l’enfant reçoit le slave comme stdio + ctty ; le parent parle via le master.
•	Resize : on relaie SIGWINCH pour garder la taille en phase (split terminal, etc.).
•	Arrêt : on forward la fin de contexte en SIGTERM, puis Kill si besoin.

Si tu veux, je te fais une variante qui monte /proc ou entre dans des namespaces (UTS/PID/mount) avant le exec, en gardant ce modèle PTY.