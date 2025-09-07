Parfait 👌 je vais te préparer un cheatsheet strace clair et structuré, pour que tu puisses le garder comme référence rapide dans tes notes.

⸻

📝 strace Cheatsheet

▶️ Lancer un programme sous strace

strace ./programme args...

→ Affiche tous les syscalls utilisés.

⸻

🔹 Options principales

Filtrer les syscalls

strace -e openat,read,write ./prog

→ Montre seulement openat, read, write.

Suivre les processus enfants (fork, clone, exec)

strace -f ./serveur

Résumé des syscalls (compte et temps passé)

strace -c ./prog

Sauvegarder la trace

strace -o trace.log ./prog

Sauvegarder par PID (utile avec -f)

strace -ff -o trace ./serveur

→ crée trace.<pid> pour chaque processus.

⸻

🔹 Attacher à un processus existant

strace -p <pid>

→ Observe en direct un programme déjà lancé.

strace -p <pid> -e trace=network

→ Ne montrer que les appels réseau (socket, connect, sendto…).

⸻

🔹 Catégories de syscalls

Tu peux filtrer par famille :

strace -e trace=file ./prog      # Fichiers
strace -e trace=process ./prog   # fork, clone, exec
strace -e trace=network ./prog   # Réseau
strace -e trace=ipc ./prog       # IPC (pipes, message queues…)
strace -e trace=memory ./prog    # mmap, brk, mprotect
strace -e trace=signal ./prog    # signaux
strace -e trace=desc ./prog      # descripteurs (read, write, close)


⸻

🔹 Suivi des fichiers et réseau

strace -e openat ./prog

→ Montre quels fichiers sont ouverts.

strace -e trace=network curl http://example.com

→ Montre toutes les connexions réseau.

⸻

🔹 Affichage avancé
•	Temps relatif entre syscalls :

strace -r ./prog


	•	Timestamps absolus :

strace -tt ./prog


	•	Timestamps haute résolution :

strace -T ./prog


	•	Inclure les valeurs de retour + erreurs :

strace -v ./prog



⸻

🔹 Trucs pratiques

Lister uniquement les syscalls utilisés (unique) :

strace -o trace.log ./prog
grep -o '^[a-z0-9_]\+' trace.log | sort -u

Tracer un seul syscall particulier :

strace -e trace=stat ./prog

Compter combien de fois chaque syscall est appelé :

strace -c ./prog


⸻

🔹 Exemple sortie typique

openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3
mmap(NULL, 123456, PROT_READ, MAP_PRIVATE, 3, 0)        = 0x7f3e8e1c0000
futex(0x7fff1234, FUTEX_WAIT, 1, NULL)                  = 0
write(1, "Hello\n", 6)                                  = 6

	•	openat → ouverture de fichier
	•	mmap → mapping mémoire
	•	futex → synchronisation threads
	•	write → écriture sur stdout

⸻

✅ En résumé :
strace = l’outil indispensable pour voir ce qu’un programme demande au noyau (fichiers, réseau, mémoire, signaux).
Avec les bonnes options (-e, -f, -c, -o), tu peux en faire un profiler, un debugger ou un générateur de profil seccomp.

⸻

Veux-tu que je t’ajoute une table des syscalls les plus fréquents (10–15 avec leur explication en une ligne) pour enrichir le cheatsheet ?