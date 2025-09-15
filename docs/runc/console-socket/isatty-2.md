Très bon point 👌 ce que tu observes est parfaitement cohérent une fois qu’on distingue les FD hérités de ton process et la présence (ou non) d’un vrai device node /dev/pts/N dans le conteneur.

⸻

🔹 Ce qui se passe dans ton conteneur
1.	Les FDs 0,1,2 pointent vers /dev/pts/0

0 -> /dev/pts/0
1 -> /dev/pts/0
2 -> /dev/pts/0

👉 Ça veut dire que ton processus a bien un terminal rattaché (un PTY esclave sur l’hôte).

	2.	Mais ton conteneur n’a pas de device nodes dans /dev

/dev/   est vide (sauf . et ..)

Donc dans l’espace de noms mount du conteneur, il n’existe pas de fichier spécial /dev/pts/0.
Pourtant, tes FDs continuent de pointer vers l’objet du noyau (le PTY) ouvert avant que le conteneur démarre.

	3.	Pourquoi tty dit “not a tty” ?
La commande tty fait en gros un isatty(0) puis appelle ttyname(0).
•	isatty(0) → vrai (fd0 est bien un tty).
•	Mais ttyname(0) essaie de résoudre vers un chemin dans /dev/…. Comme dans ton conteneur /dev/pts n’existe pas, il échoue → donc tty affiche not a tty.

⸻

🔹 Analogie
•	Tu as encore le câble branché (les FDs ouverts pointent vers le PTY).
•	Mais tu n’as plus la prise murale dans ton /dev (pas de device node pour le retrouver).

⸻

🔹 Pourquoi ça marche quand même ?
•	Les FDs ouverts sont des handles vers des objets du noyau.
•	Une fois ouverts, ils restent valides même si le chemin dans /dev disparaît ou est masqué.
•	Du coup ton process peut continuer à lire/écrire sur stdin/out/err → ça fonctionne.

⸻

🔹 Résumé
•	Ton conteneur a un terminal (via les FDs hérités).
•	Mais il n’a pas les device nodes /dev/tty, /dev/pts/*.
•	Résultat : tty (qui veut résoudre le nom du terminal via /dev) affiche “not a tty”, alors que techniquement les FDs 0/1/2 sont bien des TTY.

⸻

👉 Veux-tu que je te montre comment recréer un /dev/pts fonctionnel dans ton conteneur (avec un mount -t devpts) pour que tty redevienne capable d’afficher /dev/pts/0 ?