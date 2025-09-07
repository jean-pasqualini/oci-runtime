Parfait 👍 je vais te rédiger une fiche claire pour tes notes.

⸻

📌 strace

Qu’est-ce que c’est ?

strace est un outil de diagnostic et de débogage Linux qui permet d’observer en temps réel les appels système (syscalls) qu’un processus exécute et les signaux qu’il reçoit.

⸻

🔹 Rappel : syscalls
•	Un appel système est l’interface entre un programme utilisateur et le noyau Linux (exemple : ouvrir un fichier, allouer de la mémoire, créer un thread, envoyer un paquet réseau).
•	Exemples : read, write, openat, mmap, clone, socket, futex.

⸻

🔹 Ce que strace permet
1.	Tracer un programme : voir tous les syscalls qu’il fait.

strace ls

→ affiche tous les appels noyau utilisés par ls.

	2.	Compter et résumer (très utile pour profiler) :

strace -c ./mon_programme

→ résumé des syscalls avec leur nombre d’appels et temps passé.

	3.	Tracer les appels des sous-processus :

strace -f ./serveur

→ suit les fork/clone.

	4.	Filtrer certains syscalls seulement :

strace -e openat,read,write ./mon_programme

→ montre uniquement les ouvertures de fichiers et I/O.

	5.	Sauvegarder la trace dans un fichier :

strace -ff -o /tmp/trace ./mon_programme

→ crée /tmp/trace.<pid> pour chaque processus suivi.

⸻

🔹 Cas d’usage concrets
•	Déboguer une appli qui échoue avec Permission denied → voir quel fichier est ouvert.
•	Observer quels fichiers, bibliothèques ou sockets un programme utilise.
•	Comprendre les dépendances d’un binaire statique/dynamique.
•	Construire un profil seccomp en listant tous les syscalls utilisés.
•	Diagnostiquer des lenteurs (ex. trop de futex ou epoll_wait).

⸻

🔹 Limites
•	Overhead : ralentit un programme car chaque syscall est intercepté via ptrace.
•	Ne montre pas les appels internes aux bibliothèques (ex. appels entre fonctions Go ou C), seulement ce qui traverse vers le noyau.
•	Nécessite souvent la capacité SYS_PTRACE ou d’être root.

⸻

🔹 Alternative modernes
•	bpftrace / eBPF : plus léger, peut tracer au niveau noyau sans ralentir autant.
•	perf trace : semblable à strace, mais basé sur perf.

⸻

✅ En résumé :
strace est un stéthoscope du noyau : il écoute tout ce que ton programme demande au système d’exploitation (fichiers, mémoire, processus, réseau). C’est l’outil de base pour comprendre ce qu’un binaire fait “sous le capot”.

⸻

Veux-tu que je t’ajoute une table de correspondance entre syscalls courants et leur usage (ex: openat → ouverture fichier, futex → synchro threads) pour compléter tes notes ?


----
alternative:
auditctl and ausearch
bpftrace
seccomp-tools
SCMP_ACT_LOG
