Exact 👌, parlons du comm sous Linux.

⸻

🔹 Qu’est-ce que le comm ?
•	Chaque process a un champ “command name” dans le noyau, appelé comm.
•	Taille max = 16 caractères (stocké dans task_struct->comm).
•	Ce n’est pas forcément argv[0] (le premier argument d’execve).
•	Tu peux le voir dans :
•	/proc/<pid>/comm (juste le comm)
•	/proc/<pid>/status (Name:)
•	et utilisé par défaut par ps/top/htop.

⸻

🔹 Différence avec argv[0]
•	argv[0] → ce que l’appli a reçu comme “nom de programme”. (ex. execve("/bin/bash", ["bash", ...]))
•	comm → un champ kernel séparé, souvent initialisé à argv[0], mais qu’on peut changer avec un syscall.

Exemple :

prctl(PR_SET_NAME, "my-special-name", 0, 0, 0);

→ Ça change le comm, pas argv[0].

Donc :
•	ps -o comm= -p <pid> → affiche comm.
•	ps -o args= -p <pid> → affiche la vraie ligne de commande (argv).

⸻

🔹 Exemple en pratique

sleep 1000 &
pid=$!

# argv[0]
cat /proc/$pid/cmdline
# → "sleep\01000\0"

# comm
cat /proc/$pid/comm
# → "sleep"

Et si depuis un programme tu fais :

prctl(PR_SET_NAME, "banana", 0, 0, 0);

alors /proc/$pid/comm → "banana" mais /proc/$pid/cmdline reste "sleep 1000".

⸻

🔹 Dans le cas de runc
•	Le process enfant (init du conteneur) fait un prctl(PR_SET_NAME, "runc:[1:INIT]").
•	Résultat :
•	/proc/<pid>/comm = runc:[1:INIT]
•	/proc/<pid>/cmdline = /usr/bin/runc … init (le vrai argv[0])
•	C’est pour ça que dans ps tu vois [1:INIT] alors que c’est bien le même binaire runc.

⸻

👉 Donc le comm est juste le nom court du process, géré par le noyau, que les programmes peuvent changer (via prctl ou pthread_setname_np).

⸻

Tu veux que je te montre la différence directe avec ps (ps -o comm vs ps -o args) sur un exemple runc ?