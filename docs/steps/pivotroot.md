```
// 4) pivot_root(newRoot, newRoot/.oldroot)
	if err := unix.PivotRoot(newRoot, old); err != nil {
		return err
	}
	```

---

Short answer: no — that PivotRoot(chrootDirectory, "/") can’t work.

pivot_root(new_root, put_old) has strict rules:
	•	new_root must be a mount point (bind-mount it to itself).
	•	put_old must be a directory inside new_root (e.g. new_root/.oldroot), not /.
	•	After pivot you should chdir("/"), umount /.oldroot, and remove it.
	•	Do this inside a new mount namespace and make / rprivate first.

Here’s a minimal, correct version:

import (
	"context"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func (h *initHandler) pivotRoot(ctx context.Context) error {
	logger := logging.FromContext(ctx)

	// Choose your new root (must contain a minimal fs: /bin, /lib*, /proc, etc.)
	pwd, err := os.Getwd()
	if err != nil { return err }
	newRoot := filepath.Join(pwd, "root")
	putOld := filepath.Join(newRoot, ".oldroot")

	logger.Info("pivot_root: preparing", "newRoot", newRoot)

	// 0) Ensure we're isolated: stop mount propagation (recommended)
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return err
	}

	// 1) newRoot must be a mount point → bind-mount it to itself
	if err := unix.Mount(newRoot, newRoot, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return err
	}

	// 2) create put_old inside newRoot
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return err
	}

	// 3) pivot_root(newRoot, putOld)
	logger.Info("pivot_root: switching", "newRoot", newRoot, "putOld", putOld)
	if err := unix.PivotRoot(newRoot, putOld); err != nil {
		return err
	}

	// From here, our view of "/" is newRoot; the old root is at "/.oldroot"
	if err := os.Chdir("/"); err != nil {
		return err
	}

	// 4) detach old root and remove the directory
	if err := unix.Unmount("/.oldroot", unix.MNT_DETACH); err != nil {
		return err
	}
	if err := os.Remove("/.oldroot"); err != nil {
		return err
	}

	// 5) (re)mount /proc inside the new root
	_ = os.MkdirAll("/proc", 0555)
	if err := unix.Mount("proc", "/proc", "proc",
		uintptr(unix.MS_NOSUID|unix.MS_NODEV|unix.MS_NOEXEC), ""); err != nil {
		return err
	}

	logger.Info("pivot_root: done")
	return nil
}

Notes (important)
	•	Run this in a CLONE_NEWNS mount namespace and with CAP_SYS_ADMIN in that ns (root or rootless with NEWUSER + mappings).
	•	If you also use CLONE_NEWPID, do the /proc mount after the pivot, in the child ns, so ps reflects your PID ns.
	•	Your current function should return an error; don’t ignore failures from PivotRoot, Mount, Unmount, etc.
	•	If you only need a quick test and not full detachment, chroot(2) is simpler, but pivot_root is the proper container way because it lets you unmount the old root.

---


Parfait 👌, allons-y pas à pas pour que tu aies une bonne fiche de notes sur pivot_root.

⸻

📌 pivot_root (appel système Linux)

🔹 Définition

pivot_root(new_root, put_old) est un appel système qui permet de remplacer la racine / courante par une nouvelle racine.
C’est l’outil clé des conteneurs (runc, systemd-nspawn, chroot avancé).

⸻

🔹 Différence avec chroot
	•	chroot(new_root) : change le root apparent du processus courant et de ses enfants.
→ l’ancien root reste toujours monté, visible ailleurs.
	•	pivot_root(new_root, put_old) :
	•	Change le rootfs pour tout le mount namespace.
	•	L’ancien root est déplacé sous un répertoire (ex: /.oldroot).
	•	Tu peux ensuite le démonter → vraie isolation.

👉 pivot_root = chroot + possibilité de détacher l’ancien monde.

⸻

🔹 Conditions imposées par le noyau
	1.	new_root doit être un mount point (sinon erreur EINVAL).
→ on fait souvent un mount --bind new_root new_root.
	2.	put_old doit :
	•	être un sous-répertoire de new_root.
	•	être sur le même filesystem.
	•	être vide et disponible.
	3.	Tu dois avoir la capacité CAP_SYS_ADMIN dans le mount namespace.

⸻

🔹 Séquence typique (conteneur style runc)
	1.	Entrer dans un nouveau mount namespace (CLONE_NEWNS).
	2.	mount --make-rprivate / → casser la propagation avec l’hôte.
	3.	Préparer ton rootfs (ex: /mycontainer/rootfs).
	4.	Bind-mounter rootfs sur lui-même pour en faire un mount point.
	5.	Créer /mycontainer/rootfs/.oldroot.
	6.	Appeler :

pivot_root("/mycontainer/rootfs", "/mycontainer/rootfs/.oldroot");


	7.	Maintenant :
	•	/ pointe vers /mycontainer/rootfs.
	•	L’ancien / est accessible sous /.oldroot.
	8.	Faire chdir("/").
	9.	umount -l /.oldroot && rmdir /.oldroot.
	10.	Remonter /proc, /sys, /dev dans ce nouveau root.

⸻

🔹 Schéma

Avant :

/
├─ bin
├─ lib
├─ ...
└─ mycontainer/rootfs

Après pivot_root(rootfs, rootfs/.oldroot) :

/              → était /mycontainer/rootfs
└─ .oldroot    → l’ancien /

Puis après umount /.oldroot :

/              → rootfs propre du conteneur


⸻

🔹 Pourquoi les conteneurs l’utilisent
	•	Permet d’isoler complètement le filesystem du conteneur.
	•	Évite de garder des références à l’ancien root.
	•	Supporte la spéc OCI (rootfs configurable).
	•	Plus sûr que chroot seul, qui peut être “évadé” si des descripteurs pointent vers l’ancien root.

⸻

🔹 Exemple en Go (simplifié)

newRoot := "/path/to/rootfs"
putOld := newRoot + "/.oldroot"
unix.Mount(newRoot, newRoot, "", unix.MS_BIND|unix.MS_REC, "")
os.MkdirAll(putOld, 0700)
unix.PivotRoot(newRoot, putOld)
os.Chdir("/")
unix.Unmount("/.oldroot", unix.MNT_DETACH)
os.Remove("/.oldroot")


⸻

✅ TL;DR
	•	pivot_root = change le / courant → new_root.
	•	put_old = où basculer l’ancien /.
	•	Plus sûr que chroot car tu peux umount l’ancien monde.
	•	Obligatoire pour faire tourner un vrai conteneur rootfs isolé.

⸻

👉 Tu veux que je te fasse un diagramme ASCII étape par étape (avant pivot, après pivot, après umount) pour tes notes ?