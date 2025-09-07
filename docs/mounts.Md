Très bonne question 👌 Tu as bien compris : si on casse la propagation (MS_PRIVATE|MS_REC), on pourrait imaginer démonter tout et repartir d’un rootfs nu. Alors pourquoi garder tous les autres mounts hérités ?

⸻

🔹 Parce qu’un process a besoin d’un “système minimal” pour tourner

Même dans un conteneur, les applis s’attendent à trouver certaines choses déjà prêtes :
•	/ : ton rootfs (chroot ou pivot_root).
•	/proc : infos sur les PIDs, mémoire, CPU… sans lui, ps, top, bash ou même la libc peuvent planter.
•	/sys : infos sur les devices/kernel. Beaucoup d’outils (systemd, udev, docker…) attendent /sys.
•	/dev : accès aux TTY, /dev/null, /dev/urandom…
•	/dev/shm, /dev/mqueue : pour POSIX shared memory et message queues.
•	librairies /etc, /usr… : viennent de ton rootfs.

👉 Donc on ne démonte pas tout, sinon ton conteneur n’aurait plus rien d’utilisable.

⸻

🔹 Comment font les runtimes (runc, crun…)
1.	Clone NEWNS → nouveau mount ns.
2.	Casser la propagation → isolation (MS_PRIVATE|MS_REC /).
3.	Remonter /proc → pour refléter le PID ns.
4.	Optionnel : ajuster /sys (souvent en lecture seule).
5.	Préparer /dev : tmpfs + devpts pour les TTY.
6.	Faire pivot_root vers le rootfs du conteneur.
7.	Nettoyer l’ancien root (unmount).

Résultat : tu conserves ce qui est nécessaire, mais isolé et maîtrisé.

⸻

🔹 Pourquoi pas démonter tout par défaut ?
•	Trop brutal : tu perds /lib, /usr, /etc… qui sont nécessaires.
•	Dangereux : si tu t’y prends mal, tu peux démonter des choses critiques pour ton process et le planter.
•	Pas utile : le mount namespace t’assure déjà l’isolation, pas besoin de “purger” l’héritage, il suffit de remonter proprement les points sensibles (/proc, /sys, /dev).

⸻

🔹 Résumé
•	On garde les montages utiles (rootfs, /sys, /dev, etc.) car sinon ton conteneur est inutilisable.
•	On casse la propagation pour éviter toute fuite avec l’hôte.
•	On remplace sélectivement certains points sensibles (au minimum /proc).

⸻

👉 Veux-tu que je te fasse une liste type OCI/runc des mounts par défaut dans un conteneur Linux (ceux qu’ils gardent/remontent) ?