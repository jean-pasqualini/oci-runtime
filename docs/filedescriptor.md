Parfait 👍 allons bien dans le détail, parce que tout ton questionnement sur os.Stdout et ce que voit le parent/enfant tourne autour de la notion de file descriptor (FD).

⸻

🔹 Qu’est-ce qu’un file descriptor ?
•	C’est un entier (int) qui identifie un flux ouvert dans l’espace utilisateur.
•	Géré par le noyau Linux (et UNIX en général).
•	Chaque process a une table des descripteurs (par défaut limitée à 1024).
•	Quand tu ouvres un fichier/sock/pipe/tty, le noyau met une entrée dans sa table et te donne un petit nombre (fd).

👉 Exemple classique :
•	0 → stdin
•	1 → stdout
•	2 → stderr

⸻

🔹 Comment ça marche ?
1.	Ton programme fait open("file.txt").
2.	Le noyau crée une entrée dans sa table de fichiers ouverts (référence à un inode, offset, flags).
3.	Le noyau met un pointeur vers cette entrée dans la table de FD du process.
4.	Tu reçois un entier, ex. 3.
5.	Quand tu fais write(3, "abc"), le noyau regarde la table et écrit dans ce fichier.

⸻

🔹 FD sont hérités

Quand un process fait fork() (ou que Go lance un enfant via exec.Command), la table des FD est copiée dans l’enfant :
•	FD 0,1,2 (stdin/out/err) → pointent vers le même terminal/pipe/fichier que le parent.
•	Donc parent et enfant écrivent tous deux sur le même flux si tu ne rediriges rien.

👉 D’où ton observation : le run et le init écrivent tous les deux sur le même terminal.

⸻

🔹 os.Stdout en Go

En Go, os.Stdout est un wrapper autour de fd=1 :

fmt.Println("hello")  // écrit en fait sur os.Stdout (fd=1)
os.Stdout.Write([]byte("hello\n")) // idem, appel write(1, "hello\n")


⸻

🔹 Exemple pratique

$ echo "hello" > out.txt

	•	Le shell ouvre out.txt, obtient fd=3.
	•	Puis il fait dup2(3, 1) → ça duplique fd=3 en fd=1.
	•	Donc quand ton programme écrit sur stdout (fd=1), ça va dans out.txt.

⸻

🔹 Liens avec /proc

Dans /proc/<pid>/fd/ tu peux voir les FD d’un process :

$ ls -l /proc/$$/fd
0 -> /dev/pts/2
1 -> /dev/pts/2
2 -> /dev/pts/2

Ici, stdin, stdout, stderr pointent vers le terminal /dev/pts/2.

Si tu fais une redirection > file.txt :

1 -> /home/jean/file.txt


⸻

🔹 Résumé
•	Un file descriptor = entier qui référence un objet ouvert (fichier, socket, pipe, tty).
•	Par défaut :
•	0 = stdin
•	1 = stdout
•	2 = stderr
•	Héritage entre parent/enfant : si tu ne rediriges rien, ils écrivent tous sur le même flux.
•	En Go : os.Stdin/os.Stdout/os.Stderr sont des wrappers autour de 0,1,2.

⸻

👉 Tu veux que je t’illustre ça avec un mini-programme Go qui affiche la liste de ses FD (via /proc/self/fd) et montre à quoi ils pointent ?