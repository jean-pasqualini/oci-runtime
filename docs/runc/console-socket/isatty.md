Très bonne question 👍 isatty() est une primitive super simple mais fondamentale dans tout ce qu’on vient de discuter.

⸻

🔹 Définition

En C (POSIX) :

#include <unistd.h>
int isatty(int fd);

	•	Retourne 1 (true) si le descripteur fd est associé à un terminal.
	•	Retourne 0 (false) sinon (et met errno à ENOTTY).

⸻

🔹 Ce que ça vérifie exactement

Sous Linux, isatty(fd) est en fait un wrapper autour d’un ioctl(TCGETS) (ou équivalent plus récent TCGETS2) sur le descripteur.
•	Si l’appel réussit → le kernel dit « oui, ce FD est un périphérique de type TTY ».
•	Si ça échoue (ENOTTY) → ce n’est pas un terminal (c’est peut-être un fichier, un pipe, un socket…).

👉 En d’autres termes : isatty demande au noyau :

“ce FD est-il associé à un device qui supporte l’API termios (paramètres de terminal) ?”

⸻

🔹 Exemples

Cas 1 : interactif

$ tty
/dev/pts/3

$ isatty(0)   # stdin = terminal → vrai
$ isatty(1)   # stdout = terminal → vrai

Cas 2 : redirection

$ echo "test" > out.txt
$ ./monprog < in.txt > out.txt
isatty(0) → false   (fichier)
isatty(1) → false   (fichier)

Cas 3 : pipe

$ echo hello | ./monprog
isatty(0) → false   (pipe)
isatty(1) → true/false selon sortie (tty ou redirigée)


⸻

🔹 Utilisation pratique
•	Les shells (bash, zsh, …) utilisent isatty(0) et isatty(1) pour savoir s’ils doivent se lancer en mode interactif (prompt, job control) ou batch (script, pipe).
•	Beaucoup de programmes (ls, git log, …) utilisent isatty(1) pour décider s’ils doivent activer la colorisation :
•	stdout = terminal → afficher avec couleurs.
•	stdout = redirigé vers un fichier → pas de couleurs (juste texte brut).

⸻

✅ En résumé :
•	isatty(fd) teste si un descripteur pointe vers un device de type terminal.
•	Techniquement, ça repose sur un ioctl(TCGETS) qui n’est supporté que par les TTY/PTY.
•	Utile pour savoir si un programme tourne en mode interactif ou non.

⸻

Tu veux que je t’écrive un mini programme Go qui fait un isatty sur os.Stdin / os.Stdout et montre la différence entre tty, pipe, et fichier ?