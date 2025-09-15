Très bonne question 👌

La commande tty ne regarde pas directement /dev comme un simple ls.
Elle appelle la fonction standard POSIX ttyname(0) (donc sur le descripteur de fichier 0 = stdin).

⸻

🔹 Comment ça marche techniquement
1.	tty ouvre stdin (déjà ouvert bien sûr).
2.	Il fait un appel système isatty(0) → si false, il affiche not a tty.
3.	Sinon, il appelle ttyname(0) → qui va demander au noyau :
“Donne-moi le nom du périphérique de caractère associé à ce descripteur.”

	•	Si le device a bien un inode dans /dev/pts/N ou /dev/ttyN accessible, ttyname renvoie ce chemin.
	•	Si le noyau ne trouve aucun chemin de device associé → ttyname échoue → tty affiche not a tty.

⸻

🔹 D’où vient la confusion
•	[ -t 0 ] → appelle seulement isatty(0) → ça dit vrai si c’est un terminal (même si aucun chemin visible dans /dev).
•	tty → veut afficher un nom de fichier → dépend de la présence d’un nœud dans /dev.
•	Si /dev/pts n’est pas monté dans ton conteneur, ou si le PTY n’a pas de mapping dans /dev, alors ttyname(0) échoue → tty affiche not a tty.

⸻

⚡ Résumé
•	[ -t 0 ] = demande au noyau : est-ce un terminal ?
•	tty = demande au noyau : quel est le chemin /dev/... correspondant à ce terminal ?

Donc tu peux être dans un vrai terminal (isatty = oui), mais tty dit not a tty si /dev n’expose pas de nœud pour ce PTY.

⸻

👉 Veux-tu que je t’illustre ça avec un petit code C minimal qui montre la différence entre isatty(0) et ttyname(0) ?