Très bonne question 👌. Les deux (fork et clone) sont proches mais pas équivalents.

⸻

🔹 fork(2) (POSIX)
•	Crée un nouveau processus qui est une copie presque exacte du parent :
•	même espace mémoire (copie en COW → copy-on-write),
•	mêmes fichiers ouverts,
•	même cwd, environ, etc.
•	Parent et enfant continuent d’exécuter après l’appel.
•	Le seul vrai changement : le PID.
•	Standard POSIX, dispo sur tous les Unix.

⸻

🔹 clone(2) (Linux spécifique)
•	Syscall plus bas niveau et flexible.
•	Tu passes des flags pour choisir ce que l’enfant partage avec le parent :
•	CLONE_VM → même espace mémoire,
•	CLONE_FS → mêmes infos fs (cwd, root),
•	CLONE_FILES → même table des FDs,
•	CLONE_SIGHAND → mêmes handlers de signaux,
•	CLONE_NEWNS, CLONE_NEWPID, CLONE_NEWNET, … → créent de nouveaux namespaces,
•	etc.
•	Avec certains sets de flags, clone sert à faire :
•	un process classique (comme fork),
•	ou un thread (si tu partages mémoire, FDs, handlers).
•	C’est aussi la base des namespaces Linux → conteneurs.

⸻

🔹 Différences clés
•	fork est une version restreinte et portable : copie du process + exécution indépendante.
•	clone est Linux-only et beaucoup plus flexible : tu choisis ton degré de partage.
•	Sous Linux, fork() est en fait implémenté comme un clone() avec un ensemble de flags par défaut (genre SIGCHLD pour que le parent soit notifié quand l’enfant meurt, mais sans partage mémoire ni FDs sauf duplications).

⸻

🔹 En Go (os/exec)
•	Pour exec.Command(...), Go appelle en interne une fonction forkExec, qui sous Linux utilise clone avec les flags qui équivalent à un fork.
•	Pourquoi ? Parce que clone permet plus de contrôle, et Go a besoin de manipuler la table de FDs proprement (multi-thread safe).
•	Mais pour toi, développeur Go, ça reste sémantiquement un fork+exec classique.

⸻

⚡ Résumé
•	fork : simple, standard, copie entière du process → portable.
•	clone : outil Linux bas niveau → permet de faire fork, thread, namespace, etc. selon les flags.
•	Sous Linux : fork n’est qu’un cas particulier de clone.

⸻

👉 Veux-tu que je te montre les flags exacts que Go utilise dans son clone quand tu fais un exec.Command ?

----

Super question 👌 Tu veux voir ce que fait Go sous le capot quand tu fais :

cmd := exec.Command("ls", "-l")
cmd.Run()


⸻

🔎 Comment Go implémente exec.Command (Linux)
1.	exec.Command appelle en interne os.StartProcess.
2.	Sur Linux, os.StartProcess appelle syscall.StartProcess.
3.	Là-dedans, ça tombe sur syscall.forkAndExecInChild qui utilise clone(2) (ou vfork selon les versions/glibc).

Flags utilisés par défaut

Quand Go veut émuler un fork, il appelle clone avec les mêmes flags que fork utiliserait, c’est-à-dire :
•	SIGCHLD → pour que le parent reçoive SIGCHLD quand l’enfant meurt.
•	Pas de partage mémoire (CLONE_VM absent),
•	Pas de partage FDs (CLONE_FILES absent),
•	Pas de partage FS (CLONE_FS absent),
•	Bref : un vrai nouveau processus indépendant, comme fork.

Donc ça revient à un simple :

clone(SIGCHLD, 0);

👉 ce qui est exactement ce que fait fork() en interne sous Linux.

⸻

⚠️ Nuance importante
•	Go ne fait pas un fork() “pur” parce que dans un processus multi-thread, fork() seul est dangereux : il ne duplique qu’un seul thread → si tu touches à des locks Go après, tu peux deadlocker.
•	Donc le runtime Go a une implémentation spéciale (forkAndExecInChild) qui :
•	fait le clone(SIGCHLD),
•	nettoie l’état minimal nécessaire,
•	appelle immédiatement execve().

Il n’y a pas de retour dans l’enfant avant execve, donc pas de risque de deadlock avec le runtime.

⸻

✅ Résumé

Par défaut, exec.Command en Go fait un clone(SIGCHLD), donc sémantiquement identique à un fork() classique :
•	Nouveau process,
•	Pas de partage mémoire/FDS (sauf ceux explicitement mis dans ExtraFiles),
•	Notif SIGCHLD au parent quand l’enfant meurt,
•	Suivi immédiatement par un execve.

⸻

👉 Tu veux que je te montre un extrait du code Go (forkAndExecInChild dans src/syscall/exec_unix.go) qui montre précisément l’appel clone(SIGCHLD) ?

-----

Bonne suite 👌

🔹 CLONE_FS (flag de clone(2))
•	Si absent (cas fork) :
•	L’enfant a sa propre copie des infos “FS” (filesystem), donc indépendantes :
•	répertoire courant (cwd),
•	racine (chroot),
•	umask.
•	Quand le parent fait chdir("/tmp"), l’enfant ne bouge pas (il garde l’ancien cwd).
•	Si présent :
•	L’enfant et le parent partagent ces infos FS.
•	Si un process fait chdir, chroot ou change son umask, ça change pour l’autre aussi.

⸻

📌 Exemple simple

Sans CLONE_FS :

pid = fork();
if (pid == 0) {
chdir("/tmp");
// parent reste dans son cwd initial
}

Avec CLONE_FS :

pid = clone(CLONE_FS | SIGCHLD, 0);
if (pid == 0) {
chdir("/tmp");
// parent change aussi de cwd (ils partagent)
}


⸻

🔹 Usage typique
•	Threads (pthreads) sont créés avec CLONE_FS → tous les threads partagent cwd, root, umask.
•	Process classiques (fork) → pas de CLONE_FS → chacun son contexte FS.

⸻

👉 Donc, pour répondre à ta question :
Non, Go ne met pas CLONE_FS quand il lance un process avec exec.Command.
Il fait juste l’équivalent de fork → copie isolée du FS info.

⸻

Veux-tu que je te fasse un tableau récapitulatif des flags principaux (CLONE_VM, CLONE_FS, CLONE_FILES, CLONE_SIGHAND) avec qui est utilisé pour un process vs un thread ?

-----

Parfait 👌 On peut résumer ce que Go choisit quand tu fais un exec.Command (ou os.StartProcess) :

⸻

🔹 Pour un processus externe (exec.Command)

Go veut émuler fork + execve → donc il n’active pas les flags de partage type threads.
En pratique, ça donne :
•	SIGCHLD → obligatoire pour être notifié à la fin de l’enfant.
•	Pas de CLONE_VM → mémoire indépendante.
•	Pas de CLONE_FILES → table de FDs copiée, pas partagée.
•	Pas de CLONE_FS → cwd, root, umask copiés, pas partagés.
•	Pas de CLONE_SIGHAND → handlers de signaux indépendants.

👉 C’est donc strictement comme un fork() classique, suivi immédiatement d’un execve().

⸻

🔹 Pour un thread (runtime Go interne)

Quand le runtime Go crée un thread (ex: pour planifier des goroutines), il appelle aussi clone(), mais cette fois avec les flags de threads Linux :
•	CLONE_VM → mémoire partagée.
•	CLONE_FS → cwd/root/umask partagés.
•	CLONE_FILES → table des FDs partagée.
•	CLONE_SIGHAND → handlers de signaux partagés.
•	CLONE_THREAD → même PID que le parent (c’est ça qui fait un thread).

👉 Donc : process Go = clone façon fork, thread Go = clone façon pthread.

⸻

✅ Résumé
•	exec.Command : Go choisit uniquement SIGCHLD (isolé, comme fork).
•	Threads internes : Go choisit CLONE_VM | CLONE_FS | CLONE_FILES | CLONE_SIGHAND | CLONE_THREAD (partagé, comme pthread).

⸻

Tu veux que je te retrouve dans le code source de Go les deux endroits :
•	forkAndExecInChild (process externe),
•	newosproc (thread interne),
pour que tu voies les clone avec flags différents côte à côte ?