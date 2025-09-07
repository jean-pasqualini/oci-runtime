Voici un exemple minimal en Go qui montre “à quoi ressemble un appel à libcontainer” pour lancer un shell dans un conteneur OCI-like.
On part du principe que tu as un rootfs prêt (ex. un BusyBox rootfs dans ./rootfs).

⚠️ Teste ça sur Linux. En root c’est le plus simple. En rootless, ajoute un user namespace + UID/GID mappings.

package main

import (
"fmt"
"log"
"os"
"syscall"
"time"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
)

func main() {
// 1) Où libcontainer va stocker ses états (state, fifo, …)
root := "/tmp/libcontainer-demo"
if err := os.MkdirAll(root, 0o755); err != nil {
log.Fatal(err)
}

	// 2) Fabrique libcontainer (backend cgroups : Cgroupfs ou Systemd)
	factory, err := libcontainer.New(root, libcontainer.Cgroupfs)
	if err != nil {
		log.Fatal(err)
	}

	// 3) Définition de la config du conteneur
	cfg := &configs.Config{
		Rootfs: "rootfs", // chemin vers le rootfs (ex. ./rootfs)
		// Namespaces à créer (le strict minimum : PID + MOUNT + UTS + IPC)
		Namespaces: configs.Namespaces{
			{Type: configs.NEWNS},
			{Type: configs.NEWPID},
			{Type: configs.NEWUTS},
			{Type: configs.NEWIPC},
			// {Type: configs.NEWNET}, // optionnel (nécessite config réseau)
			// {Type: configs.NEWUSER}, // rootless → nécessite UID/GID mappings
		},
		// Cgroups (exemple minimal)
		Cgroups: &configs.Cgroup{
			Name:   "demo-cgroup",
			Parent: "system.slice", // ou vide si cgroupfs pur; adapte selon ta machine
			Resources: &configs.Resources{
				PidsLimit: 128,
			},
		},
		// Montages de base : /proc, /dev (tmpfs)…
		Mounts: []*configs.Mount{
			{
				Device:      "proc",
				Destination: "/proc",
				Source:      "proc",
				Flags:       syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV,
			},
			{
				Device:      "tmpfs",
				Destination: "/dev",
				Source:      "tmpfs",
				Flags:       syscall.MS_NOSUID | syscall.MS_STRICTATIME,
				Data:        "mode=755",
			},
			// (optionnel) /dev/shm
			{
				Device:      "tmpfs",
				Destination: "/dev/shm",
				Source:      "shm",
				Flags:       syscall.MS_NOSUID | syscall.MS_NODEV | syscall.MS_NOEXEC,
				Data:        "mode=1777,size=65536k",
			},
		},
		// Capabilities minimales (exemple : rien de spécial ici → dépend de ton besoin)
		Capabilities: &configs.Capabilities{
			Bounding:    []string{},
			Effective:   []string{},
			Inheritable: []string{},
			Permitted:   []string{},
			Ambient:     []string{},
		},
		// Mask/Readonly paths usuels (comme runc)
		MaskPaths: []string{
			"/proc/kcore",
			"/proc/latency_stats",
			"/proc/timer_list",
			"/proc/sched_debug",
		},
		ReadonlyPaths: []string{
			"/proc/asound",
			"/proc/bus",
			"/proc/fs",
			"/proc/irq",
			"/proc/sys",
			"/proc/sysrq-trigger",
		},
		NoNewPrivileges: true,
		Hostname:        "demo",
	}

	// 4) (Optionnel) Rootless : mapper l’UID/GID hôte -> 0 dans le userns
	// if !isRoot() {
	// 	cfg.Namespaces = append(cfg.Namespaces, configs.Namespace{Type: configs.NEWUSER})
	// 	uid := os.Getuid()
	// 	gid := os.Getgid()
	// 	cfg.UidMappings = []configs.IDMap{{HostID: uid, ContainerID: 0, Size: 1}}
	// 	cfg.GidMappings = []configs.IDMap{{HostID: gid, ContainerID: 0, Size: 1}}
	// 	cfg.GidMappingsEnableSetgroups = false
	// }

	// 5) Créer le conteneur
	containerID := fmt.Sprintf("demo-%d", time.Now().Unix())
	container, err := factory.Create(containerID, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer container.Destroy() // nettoyage cgroups/états

	// 6) Définir le process à lancer (sera le PID 1 dans le NEWPID)
	process := &libcontainer.Process{
		Args:   []string{"/bin/sh"}, // doit exister dans ton rootfs
		Env:    []string{"PATH=/bin:/sbin:/usr/bin:/usr/sbin"},
		Cwd:    "/",
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		// ConsoleWidth/Height possibles si TTY; ici on utilise stdio du parent
	}

	// 7) Démarrer (montages, pivot_root, seccomp/caps, puis execve("/bin/sh"))
	if err := container.Run(process); err != nil {
		log.Fatal(err)
	}

	// 8) Attendre la fin du shell
	_, err = process.Wait()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conteneur terminé.")
}

func isRoot() bool { return os.Geteuid() == 0 }

Préparer rapidement un rootfs BusyBox (exemple)

# BusyBox statique (exemple tout-en-un)
mkdir -p rootfs/bin rootfs/dev rootfs/proc rootfs/tmp
curl -L -o rootfs/bin/busybox https://busybox.net/downloads/binaries/1.35.0-i686-uclibc/busybox
chmod +x rootfs/bin/busybox
( cd rootfs/bin && ln -sf busybox sh && ln -sf busybox ls && ln -sf busybox ps )
# Tu peux ajouter d’autres liens (cat, echo, mount, etc.) selon tes besoins.

Points clés à retenir
•	libcontainer.New(...) → fabrique (backend cgroups).
•	factory.Create(id, cfg) → crée un conteneur avec ta configs.Config.
•	container.Run(process) → effectue namespaces + mounts + pivot_root + seccomp/caps puis execve.
•	Ton /bin/sh voit son propre /proc (grâce à NEWNS + montage proc) et ses propres PIDs (grâce à NEWPID).

Si tu veux une version rootless prête à l’emploi (avec NEWUSER + mappings) ou l’ajout du namespace réseau avec lo up, dis-moi et je te la rédige.


----


Yes 🙌, là on rentre dans les entrailles de runc.
Ce que tu soulèves, c’est le mécanisme de reexec dans runc (et plus généralement dans plein d’outils Go comme Docker, containerd, etc.).

⸻

🔹 Pourquoi un reexec ?
•	runc est un seul binaire, mais il a deux “rôles” :
1.	Parent CLI : lire la spec, préparer les cgroups, créer les namespaces.
2.	Init du conteneur : s’exécuter à l’intérieur du nouveau contexte, pour configurer le rootfs, monter /proc, puis execve le vrai binaire utilisateur.
•	Ces rôles sont séparés par un reexec : le parent se clone/fork, puis relance le même binaire avec un autre chemin de code (pas la CLI normale).

⸻

🔹 Comment c’est fait dans le code Go ?

Dans libcontainer/utils/reexec tu trouves un mécanisme qui permet d’enregistrer des handlers :

// Exemple simplifié
func init() {
reexec.Register("init", func() {
// Code spécial qui ne fait PAS la CLI
// → c’est le libcontainer init
libcontainerInit()
})
}

Et dans main.go de runc :
•	Si le process démarre “normalement”, il passe par la CLI (urfave/cli).
•	Mais si le process a été lancé avec le nom spécial init (via reexec), alors ça ne lance pas la CLI mais directement la fonction associée.

⸻

🔹 Concrètement dans runc
•	Parent runc appelle un clone() dans les bons namespaces.
•	Cet enfant fait un reexec.Self() avec l’argument init.
•	Donc au lieu de repartir dans la CLI, le binaire relancé exécute la fonction enregistrée sous "init".
•	C’est ça qui donne le runc:[1:INIT] que tu vois dans ps.

⸻

🔹 Exemple ultra simplifié

func main() {
// cas normal
if reexec.Init() {
// Si on est en mode reexec → ne pas continuer dans la CLI
return
}

    // Sinon → code CLI classique
    fmt.Println("Ceci est la CLI de runc-like")
}

func init() {
reexec.Register("init", func() {
fmt.Println("Hello, je suis le process init du conteneur")
// Ici : montage rootfs, seccomp, pivot_root, puis execve
os.Exit(0)
})
}

Puis le parent peut lancer :

reexec.Command("init").Run()

→ ça relance le même binaire avec le mode "init" enregistré.

⸻

🔹 Ce que ça apporte
•	Séparation claire : pas de gros if dans tout le code, juste un handler spécial.
•	Même binaire : pas besoin d’installer un autre exécutable genre tini.
•	Simplicité Go : le package reexec est un helper pratique, utilisé aussi dans dockerd.

⸻

👉 En résumé :
Le fameux runc:[1:INIT] est en fait runc relancé par lui-même, mais en mode init via le mécanisme reexec.
C’est du code interne à libcontainer qui fait le setup du conteneur, avant de execve ton process final.

⸻

Veux-tu que je t’ouvre le chemin dans le repo GitHub de runc (libcontainer/nsenter, reexec, init_linux.go) pour que tu voies la partie exacte où ce reexec est fait ?

---

prctl(PR_SET_NAME, "runc:[1:INIT]", 0, 0, 0)

---


🔹 Caractéristiques de libcontainer
•	Langage : Go
•	Appels système : utilise golang.org/x/sys/unix et parfois syscall.RawSyscall pour appeler directement les syscalls (clone, prctl, setns, pivot_root, mount, capset, seccomp via prctl + BPF, etc.).
•	Organisation :
•	libcontainer/configs → description d’un conteneur (namespaces, cgroups, mounts…).
•	libcontainer/cgroups → code Go qui manipule /sys/fs/cgroup/*.
•	libcontainer/nsenter → code qui utilise setns() pour rejoindre un namespace existant.
•	libcontainer/init_linux.go → le fameux code du process init.
•	libcontainer/utils/reexec → gestion du reexec dont on a parlé.