Très bonne intuition 👍 : on associe souvent job control au shell (bash, zsh…), mais en fait ce n’est pas une invention du shell.

C’est une mécanique du kernel Unix qui permet de contrôler plusieurs process groups rattachés à un terminal (ctty). Le shell est juste l’utilisateur le plus courant de cette mécanique.

⸻

🔹 Le job control en quelques points
1.	Process groups
•	Chaque process appartient à un process group (identifié par un PGID).
•	Le kernel gère ça pour savoir quels process sont ensemble (ex : vim et son enfant less dans le même job).
2.	Foreground vs Background
•	Un terminal (ctty) n’a qu’un seul process group de foreground.
•	Les autres process groups de la session sont en background.
•	Le kernel envoie les signaux clavier (SIGINT, SIGTSTP, SIGQUIT, SIGWINCH) uniquement au foreground process group.
3.	Le rôle du shell
•	Quand tu tapes sleep 1000 &, bash crée un nouveau process group pour sleep et le marque en background.
•	Quand tu tapes fg %1, bash demande au kernel de mettre ce process group au foreground → il reprend le terminal et recevra Ctrl-C, Ctrl-Z.
•	Quand tu tapes Ctrl-Z, le kernel envoie SIGTSTP au foreground group → bash peut ensuite “suspendre” ce job et l’envoyer en background avec bg.

⸻

🔹 Exemple en schéma

TTY (/dev/pts/3) ← ctty de la session
|
v
Foreground group:  [vim] (PGID=1234)
Background group:  [sleep 1000] (PGID=5678)

Ctrl-C  → SIGINT envoyé à PGID=1234
Ctrl-Z  → SIGTSTP envoyé à PGID=1234
(sleep en background n’est pas affecté)


⸻

🔹 Conclusion
•	Le job control est une fonctionnalité du noyau Unix liée aux concepts :
•	sessions
•	process groups
•	controlling terminal (ctty)
•	bash/zsh ne font que fournir une interface utilisateur :
•	%1, fg, bg, jobs…
•	Eux-mêmes utilisent des appels système (setpgid, tcsetpgrp, etc.) pour manipuler les process groups.

⸻

👉 Donc : bash ≠ job control ; bash utilise le job control du noyau pour te donner les commandes pratiques.

Veux-tu que je te montre la liste des appels systèmes précis que le shell utilise pour implémenter jobs, fg, bg ?

----

Yes ✅ tu tiens le bon fil, je t’affine un peu les nuances :

⸻

🔹 1. Qui décide du foreground process group ?
•	C’est le session leader (typiquement ton shell interactif) qui a le ctty (controlling terminal).
•	Lui seul peut appeler tcsetpgrp(ctty_fd, pgid) pour dire au kernel :
« À partir de maintenant, ce process group est au foreground ».
•	Exemple :
•	Bash → leader de session, ctty = /dev/pts/3.
•	Quand tu tapes fg %1, bash fait tcsetpgrp(tty_fd, pgid_du_job) pour donner le terminal au bon job.

⸻

🔹 2. Qui reçoit un “vrai terminal interactif” ?
•	Le process group de foreground → reçoit :
•	les signaux clavier (Ctrl-C, Ctrl-Z, …),
•	les événements de resize (SIGWINCH),
•	l’accès exclusif à la lecture de ce TTY (les autres process groups qui lisent dessus reçoivent EIO ou sont bloqués).
•	Les autres process groups en background :
•	N’ont pas d’interaction directe avec le terminal.
•	Si un process background essaie de lire depuis le terminal → le kernel lui envoie SIGTTIN.
•	S’il essaie d’écrire → SIGTTOU.
•	Cela empêche deux jobs d’interférer sur le même terminal en même temps.

⸻

🔹 3. Le rôle du ctty
•	Seule la session qui possède le ctty a droit au “terminal interactif”.
•	Les autres sessions (sans ctty) ne peuvent pas recevoir ces signaux → ce sont des daemons ou jobs batch.

⸻

🔹 Exemple schéma ASCII

/dev/pts/3 (ctty)
|
v
Session leader = [bash]  (a ouvert le TTY, fait setsid + TIOCSCTTY)

Process groups :
┌──────────────────────────────┐
│ Foreground group = [vim]     │  ← reçoit signaux clavier
└──────────────────────────────┘
┌──────────────────────────────┐
│ Background group = [sleep 100] │  ← SIGTTIN si lit depuis TTY
└──────────────────────────────┘


⸻

🔹 Résumé
•	✔️ Oui, c’est bien le session leader qui choisit quel process group est au foreground → et donc qui a le “vrai terminal interactif”.
•	✔️ Les autres process groups de la même session existent, mais ils n’ont pas accès directement au terminal interactif (ils reçoivent des signaux SIGTTIN/SIGTTOU s’ils essaient).
•	✔️ Les sessions sans ctty (daemons, conteneurs détachés, cron jobs…) n’ont pas de terminal interactif du tout.

⸻

👉 Veux-tu que je te montre un petit exemple en bash pur où tu vois clairement SIGTTIN/SIGTTOU quand un job background essaie de lire/écrire sur le terminal ?