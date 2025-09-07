Parfait 👍 je te fais un cheatsheet compact qui résume ce que fait runc quand il lance un conteneur. C’est l’enchaînement classique en mode rootful (et j’indique les différences rootless).

⸻

📝 Cheatsheet : Cycle de vie d’un conteneur avec runc

1. Lecture de la spec OCI
   •	Fichier config.json : décrit :
   •	root.path → le rootfs (système de fichiers de base).
   •	process.args → la commande à exécuter (bash, sleep, etc.).
   •	mounts → /proc, /sys, /dev, volumes, tmpfs…
   •	linux.namespaces → liste des ns à créer (pid, mount, uts, ipc, net, user, cgroup).
   •	linux.resources → cgroups (CPU, mémoire, IO, pids…).
   •	linux.seccomp / AppArmor / SELinux / capabilities.

⸻

2. Création du processus init
   •	runc fait un clone(CLONE_NEW*) avec les flags demandés.
   •	Typiquement :
   •	CLONE_NEWPID → PIDs isolés.
   •	CLONE_NEWNS → montages isolés.
   •	CLONE_NEWUTS → hostname.
   •	CLONE_NEWIPC → shm, semaphores.
   •	CLONE_NEWNET → interfaces réseau.
   •	CLONE_NEWUSER (rootless) → UID/GID mappés.
   •	Le nouveau process devient PID 1 du conteneur.

⸻

3. Préparation du rootfs
   •	Monte le rootfs défini dans config.json.
   •	Applique pivot_root() → le root du conteneur devient rootfs/.
   •	Rend les montages privés (MS_PRIVATE) pour ne pas propager vers l’hôte.
   •	Monte les systèmes nécessaires :
   •	/proc (proc) → reflète le PID ns.
   •	/sys (sysfs) → souvent en lecture seule.
   •	/dev (tmpfs + mknod ou bind de tty, null, zero, random).
   •	/dev/shm, /dev/mqueue, etc.
   •	Applique les bind-mounts et volumes.

⸻

4. Configuration des cgroups
   •	Parent runc place le PID du conteneur dans les cgroups.
   •	Applique limites (CPU, mémoire, IO, pids.max).
   •	Rootless → cgroups restreints (selon ce qui est autorisé).

⸻

5. Sécurité et restrictions
   •	capabilities → drop tout sauf un set minimal (NET_BIND_SERVICE, SYS_CHROOT…).
   •	no_new_privs=1.
   •	seccomp → filtre BPF des syscalls interdits.
   •	AppArmor/SELinux profil si activé.
   •	maskedPaths (ex: /proc/kcore) et readonlyPaths (ex: /proc/sysrq-trigger).

⸻

6. Lancement du processus utilisateur
   •	Le runc:[1:INIT] (processus init du conteneur) fait un execve() du process.args.
   •	Ce process devient PID 1 du conteneur.
   •	runc garde un canal de communication pour kill, exec, delete…

⸻

7. Fin de vie
   •	Quand le PID 1 du conteneur meurt → le conteneur s’arrête.
   •	runc nettoie : cgroups, namespaces, descripteurs…

⸻

🧩 Résumé visuel

runc (hôte)
│
├─> clone(CLONE_NEWPID | NEWNS | UTS | IPC | NET [| USER])
│       │
│       └─ init process (PID 1 du conteneur)
│             ├─ pivot_root() -> rootfs/
│             ├─ mount /proc, /sys, /dev
│             ├─ appliquer seccomp/caps
│             └─ execve("bash")
│
└─ configure cgroups + surveille le conteneur


⸻

👉 Tu veux que je t’en fasse une version tableau avec les colonnes : namespace / isolement / comment runc l’utilise ?

----

Yes 👌, exactement : libcontainer est le cœur technique de runc.

⸻

🔹 C’est quoi libcontainer ?
•	Une librairie Go (dans le repo opencontainers/runc/libcontainer).
•	Initialement développée par Docker (2014) → a ensuite été extraite quand l’OCI (Open Container Initiative) est née.
•	Elle gère tout le bas niveau des conteneurs en parlant directement au kernel Linux (via syscall, unix package, cgroups, namespaces).
•	Pas de dépendance à LXC : au contraire, libcontainer a été créée pour se débarrasser de la dépendance historique de Docker à LXC.

⸻

🔹 Rôles de libcontainer
1.	Création des namespaces
•	Via clone/clone3 avec les bons flags (CLONE_NEWPID, CLONE_NEWNS, CLONE_NEWNET, …).
•	Support rootless : création de userns + mapping UID/GID.
2.	Montages et rootfs
•	pivot_root, chroot, bind-mounts.
•	Montage de /proc, /sys, /dev, volumes.
3.	Cgroups
•	Crée, configure et attache les processus aux cgroups (v1/v2).
4.	Sécurité
•	Applique seccomp (filtres BPF).
•	Configure AppArmor/SELinux.
•	Gère capabilities (via capset).
•	Met no_new_privs.
5.	Init process
•	Code de l’init du conteneur (le fameux runc:[1:INIT]).
•	Gestion des signaux + reaper des zombies.
•	execve() du vrai process utilisateur.
6.	Interaction avec runc
•	runc est juste une CLI qui appelle libcontainer.
•	Toute la logique “run, exec, kill, delete” passe par les API de libcontainer.

⸻

🔹 Architecture simplifiée

┌────────────┐
│  runc CLI  │  ← (outil en ligne de commande, lit config.json OCI)
└─────┬──────┘
│ utilise
┌─────▼──────────────────┐
│      libcontainer       │  ← (lib Go qui parle directement au kernel)
│  • namespaces           │
│  • cgroups              │
│  • mounts / pivot_root  │
│  • seccomp / LSM        │
│  • init process (PID 1) │
└───────────┬────────────┘
│ syscalls
┌────▼─────┐
│  kernel  │  ← (namespaces, cgroups, capabilities…)
└──────────┘


⸻

🔹 Exemple concret

Quand tu fais :

runc run myctr

	•	runc (CLI) lit le config.json.
	•	Il appelle libcontainer pour :
	•	libcontainer.New() → prépare la config Go.
	•	container.Start() → clone les ns, setup cgroups, mount rootfs, seccomp.
	•	container.Exec() → lance le process demandé.

⸻

👉 Donc : le petit process init dont on parlait plus tôt est directement du code de libcontainer (dans init_linux.go et init.go), pas un binaire externe.

⸻

Tu veux que je te montre un bout de code source Go de libcontainer qui correspond au lancement de ce process init ?