Parfait — voici une checklist des syscalls essentiels pour un runtime de conteneur, classés par brique fonctionnelle, avec l’usage typique et (quand utile) l’ordre logique d’appel.

Isolation par namespaces
•	clone/clone3 (CLONE_NEW*) — créer le process init du conteneur avec les namespaces : CLONE_NEWNS, CLONE_NEWPID, CLONE_NEWUTS, CLONE_NEWNET, CLONE_NEWIPC, CLONE_NEWUSER, CLONE_NEWCGROUP (si dispo).
•	unshare (CLONE_NEW*) — séparer le namespace du process courant (utile si tu veux un modèle “unshare + fork”).
•	setns — joindre un namespace existant (ex. pour exec latéral, attach, nsenter-like).
•	sethostname — configurer l’UTS namespace (nom d’hôte du conteneur).

Identités & privilèges (userns)
•	setresuid / setresgid / setgroups — basculer vers l’UID/GID voulu dans le conteneur.
•	(via FS) écriture dans /proc/<pid>/{uid_map,gid_map,setgroups} — établir les mappings UID/GID en userns (implique openat, write).
•	capset / capget — définir/inspecter les Linux capabilities (drop au minimum nécessaire).
•	prctl(PR_SET_NO_NEW_PRIVS) — empêcher l’élévation de privilèges future.
•	prlimit64 / setrlimit — appliquer des limites de ressources par process.

Cgroups (v2 ou v1)
•	openat / mkdirat / write sur /sys/fs/cgroup/… — créer le cgroup, attacher le PID, configurer cpu.max, memory.max, pids.max, etc.
•	pidfd_open / pidfd_send_signal — pratique pour gérer/terminer proprement le PID1 du conteneur.

Système de fichiers (rootfs)
•	mount / umount2 — monter le rootfs, bind-mounts, proc, sys, dev, propagation (MS_SLAVE/MS_PRIVATE).
•	pivot_root — basculer vers la nouvelle racine (puis démonter l’ancienne).
•	chdir / chroot — chdir("/") après pivot_root; chroot possible selon design (moins moderne que pivot_root).
•	openat / openat2 / newfstatat / statx — ouvrir/stat-er en relatif (sécurisé), vérifier la présence des fichiers.
•	mkdirat / mknodat / linkat / symlinkat / renameat2 / unlinkat — préparer l’arborescence du conteneur.
•	fchmodat / fchownat / utimensat — ajuster permissions/propriétés si besoin.

API de montages moderne (si tu cibles les noyaux récents) :

	•	fsopen / fsconfig / fsmount / move_mount / open_tree / fspick — nouvelle API de montages plus sûre/atomique.

Réseau
•	socket / bind / listen / accept4 / connect — sockets applicatifs.
•	setsockopt / getsockopt — options réseau (ex. SO_REUSEPORT, SO_MARK).
•	ioctl(SIOCGIF*) — ops réseau legacy (config d’interface si tu gères ça en userspace).
•	(Netlink) socket(AF_NETLINK) + sendmsg/recvmsg — configurer interfaces, routes, veth, bridges via rtnetlink dans le netns.

Sécurité (LSM, seccomp, etc.)
•	prctl(PR_SET_SECCOMP, SECCOMP_MODE_FILTER, …) ou seccomp — installer ton filtre BPF seccomp.
•	bpf — si tu compiles/charges des programmes BPF (pour seccomp avancé ou observabilité).
•	keyctl / écriture dans /proc/self/attr/* — interactions LSM/AppArmor/SELinux selon besoin.

Exécution & supervision
•	execve / execveat — lancer le binaire du conteneur.
•	wait4 / waitid — attendre le PID1 du conteneur, propager le code retour.
•	kill / tgkill / pidfd_send_signal — relayer signaux (SIGTERM/SIGKILL) vers le conteneur.
•	signalfd4 / eventfd2 / timerfd_create / epoll_create1 + epoll_ctl/epoll_wait — boucle d’événements pour PID1/supervision.
•	clone3 (avec CLONE_PARENT_SETTID/CLONE_CHILD_CLEARTID) — gestion fine de TIDs.

Mémoire, temps, entropie (courants mais utiles)
•	mmap / munmap / mprotect / madvise / brk — allocations et protections mémoire.
•	futex — synchro rapide (threads).
•	nanosleep / clock_gettime — temporisation/mesure.
•	getrandom — entropie pour seeds/IDs.

Utilitaires de base (FS & infos)
•	read / write / pread64 / pwrite64 / close / fcntl / dup/dup3 / pipe2 — I/O classiques et pipes.
•	ioctl — divers contrôles sur FDs (tty, net, etc.).
•	getpid / getppid / gettid / uname / getcwd — introspection basique.

⸻

Ordre type (résumé « happy path »)
1.	clone3/unshare avec CLONE_NEWUSER (puis mapper UID/GID) → autres namespaces (NEWNS, NEWUTS, NEWIPC, NEWNET, NEWPID).
2.	sethostname (UTS).
3.	Rootfs : mount des bind, proc, sys, dev → pivot_root → chdir("/") → nettoyage des anciens montages.
4.	Cgroups : créer le cgroup, attacher le PID, appliquer limites.
5.	Capabilités & no_new_privs : capset, prctl(PR_SET_NO_NEW_PRIVS).
6.	Seccomp : installer le filtre.
7.	drop privileges : setresuid/setresgid/setgroups (ordre crucial), umask si besoin.
8.	execve de l’entrée du conteneur.
9.	supervision du PID1 depuis le parent : waitid, relais de signaux, collecte du code retour.

⸻

Si tu veux, je peux te générer une matrice “syscall → feature” (colonnes : nécessaire/recommandé/optionnel, namespace concerné, notes kernel) pour l’ajouter direct à ton repo en Markdown.