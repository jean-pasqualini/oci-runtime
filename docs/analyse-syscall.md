Parfait 👍
Je vais t’expliquer le rôle de chaque syscall trouvé dans ta trace. Certains sont très courants, d’autres plus spécialisés. Voici les descriptions :

⸻

Gestion de processus / exécution
	•	execve : lance un programme en remplaçant l’image du processus courant par une nouvelle.
	•	clone / clone3 : crée un nouveau processus ou thread (utilisé pour fork, pthread, etc.).
	•	exit_group : termine tous les threads du processus.
	•	waitid : attend la fin d’un processus enfant.
	•	tgkill : envoie un signal à un thread spécifique.
	•	rt_sigaction : définit ou lit le gestionnaire d’un signal.
	•	rt_sigprocmask : bloque ou débloque certains signaux.
	•	rt_sigreturn : retourne d’un handler de signal.
	•	sigaltstack : définit une pile alternative pour gérer les signaux.
	•	prctl : configure divers paramètres de processus (nom, comportement, etc.).
	•	set_tid_address / set_robust_list / rseq : utilisés pour la gestion des threads et des verrous.

⸻

Gestion mémoire
	•	mmap : mappe un fichier ou une zone mémoire dans l’espace du processus.
	•	munmap : libère une zone mémoire précédemment mappée.
	•	mprotect : change les permissions d’une zone mémoire.
	•	brk : ajuste la fin du segment de données (allocation mémoire simple).
	•	madvise : donne des conseils au noyau sur l’usage prévu d’une zone mémoire.

⸻

Fichiers & répertoires
	•	openat : ouvre un fichier ou un répertoire (relative à un descripteur de dir).
	•	read / pread64 : lit des données dans un fichier.
	•	write : écrit dans un fichier ou un descripteur.
	•	close : ferme un descripteur de fichier.
	•	unlinkat : supprime un fichier.
	•	mkdirat : crée un répertoire.
	•	newfstatat / fstat : récupère des infos sur un fichier (taille, permissions…).
	•	faccessat : vérifie les droits d’accès à un fichier.
	•	readlinkat : lit le contenu d’un lien symbolique.
	•	getcwd : renvoie le chemin du répertoire courant.

⸻

Descripteurs de fichiers
	•	fcntl : contrôle divers paramètres de fichiers (flags, locks, etc.).
	•	dup / dup3 : duplique un descripteur de fichier.
	•	pipe2 : crée un pipe (communication entre processus).

⸻

Réseau & sockets
	•	socket : crée un point de communication réseau.
	•	connect : établit une connexion vers un socket.

⸻

Système / noyau
	•	sched_getaffinity : récupère les CPUs sur lesquels un processus peut s’exécuter.
	•	uname : récupère les infos système (nom noyau, version, architecture).
	•	prlimit64 : fixe ou lit les limites de ressources (ex. mémoire max, nombre de fichiers).
	•	mount / umount2 : monte ou démonte un système de fichiers.
	•	pivot_root : change la racine du système de fichiers (utilisé dans les conteneurs).
	•	sethostname : change le nom d’hôte du système.
	•	ioctl : interface générique pour contrôler un périphérique.
	•	getrandom : génère des nombres aléatoires du noyau.

⸻

Identité / permissions
	•	getpid : renvoie l’ID du processus.
	•	getppid : renvoie l’ID du parent.
	•	gettid : renvoie l’ID du thread.
	•	getuid / geteuid : renvoie l’UID réel / effectif.
	•	getgid / getegid : renvoie le GID réel / effectif.
	•	setpgid : change le groupe de processus (utile pour les jobs shell).
	•	getpgid : lit le groupe de processus.

⸻

Synchronisation & timing
	•	futex : mécanisme de synchronisation rapide (mutex/condvar).
	•	nanosleep : met un processus en pause pour une durée donnée.
	•	pselect6 : attend la disponibilité d’E/S avec masquage de signaux.

⸻

Spéciaux / erreurs
	•	EINVAL : n’est pas un syscall → c’est un code d’erreur “Invalid argument”.
	•	EAGAIN : idem, code d’erreur “Try again”.

⸻

👉 Donc, certains (EINVAL, EAGAIN) ne sont pas des syscalls mais des erreurs affichées dans le log.

Veux-tu que je fasse un tableau récapitulatif clair (syscall → description) pour que ce soit plus lisible et exportable ?