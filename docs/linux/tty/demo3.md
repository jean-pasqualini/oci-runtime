Ahh, tu viens de mettre le doigt sur la subtilité clé 👌

👉 Ce ne sont pas les programmes qui “changent de PGID” en permanence.
Chaque processus a un PGID fixe (sauf s’il crée volontairement un nouveau groupe avec setpgid() ou setsid()).

Ce qui change dynamiquement, c’est :

🔑 Le foreground process group du TTY
•	Chaque terminal (/dev/pts/0) garde un identifiant : “le PGID qui a le droit de lire mon clavier en ce moment”.
•	Ce choix est manipulé par le shell (via tcsetpgrp()), pas par les programmes eux-mêmes.

⸻

🎭 Rôles des différents acteurs
•	Process group (PGID)
•	Sert à grouper les processus d’un “job” (ex : make, son go run, etc.).
•	Stable une fois créé, sauf si un process appelle explicitement setpgid() ou setsid().
•	Foreground process group (tty → PGID)
•	Unique : il n’y a qu’un seul PGID en foreground par terminal.
•	Peut changer quand le shell décide de donner la main à un job (fg, lancement normal d’une commande).
•	Tous les autres PGID rattachés au même TTY deviennent background process groups.

⸻

📌 Exemple concret
1.	Tu es dans bash (PGID=1).
→ Foreground PGID = 1.
→ bash reçoit ton clavier.
2.	Tu lances make run-shim.
•	bash met PGID=7 au foreground avec tcsetpgrp.
•	Foreground PGID = 7.
•	bash passe en attente → il ne lit plus ton clavier.
3.	Le job lance /bin/sh -l dans un nouveau groupe (PGID=122).
•	Foreground PGID bascule à 122.
•	Maintenant, c’est /bin/sh -l qui reçoit le clavier.

⸻

✅ Donc :
•	Les processus ne changent pas de PGID à la volée.
•	C’est le TTY qui élit un PGID “au micro” → le foreground group.
•	Les autres PGID du même terminal restent en background.

⸻

Veux-tu que je te fasse un petit schéma ASCII avec bash, make, leurs PGID et la flèche qui montre “qui est foreground” ? Ça rend souvent le truc beaucoup plus clair visuellement.