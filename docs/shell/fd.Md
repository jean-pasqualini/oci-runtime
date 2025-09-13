Exact, tu as bien repéré 👌

🔹 Le fameux FD 255 de bash
•	Bash ouvre son propre descripteur haut (souvent 255) dès le démarrage.
•	Ce FD est un handle supplémentaire sur le TTY de contrôle (pas un pipe).
•	Bash s’en sert pour :
•	garder un accès direct au terminal même si l’utilisateur redirige stdin/stdout/stderr,
•	restaurer l’état du terminal (termios, modes de ligne) après exécution de commandes interactives,
•	gérer le job control (SIGTTOU, fg, bg, etc.).

En gros : FD 255 = copie de secours sur le terminal que Bash garde tout le temps ouvert.

⸻

🔹 Pourquoi pas un pipe ?
•	Bash a besoin de parler au vrai terminal :
•	pour la lecture interactive (read -e, édition de ligne),
•	pour manipuler les attributs termios (désactiver/activer echo, modes raw/canonique),
•	pour envoyer/recevoir des signaux de job control.
•	Un pipe anonyme ne permet ni termios, ni d’être un TTY de contrôle.
•	Donc il doit ouvrir /dev/tty (qui pointe vers le terminal associé au process).

⸻

🔹 Autres shells
•	dash / ash ouvrent aussi parfois un FD supplémentaire (souvent 10).
•	Zsh a ses propres FDs internes pour job control.
•	Bref : chaque shell interactif garde une poignée persistante sur /dev/tty.

⸻

🔹 Résumé
•	FD 255 dans Bash ≠ fuite → c’est volontaire.
•	Il pointe sur /dev/tty (ton pty) pour que Bash garde toujours une ligne directe avec le terminal, même si tu fais des redirections folles.
•	Impossible avec un simple pipe, car il faut un vrai device TTY pour la gestion interactive.

⸻

👉 Donc : pas besoin de t’en inquiéter ni de le fermer dans ton runtime. C’est du domaine du shell, pas une fuite de ton côté.

Veux-tu que je te montre en strace le moment exact où Bash ouvre FD 255 au lancement ?