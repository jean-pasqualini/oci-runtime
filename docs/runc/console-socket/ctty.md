
Très bonne question 👌

La réponse courte : non, un terminal (TTY/PTY) ne peut être le controlling terminal (ctty) que d’une seule session à la fois.

⸻

🔹 Règles du kernel (simplifiées)
1.	Un terminal ↔ au plus une session
•	Quand un processus fait setsid() puis ioctl(TIOCSCTTY), le terminal devient le ctty de cette session.
•	Si une autre session essaie de prendre ce même terminal :
•	par défaut, TIOCSCTTY échoue avec EPERM,
•	sauf si on utilise le mode force (rare, réservé à init/getty pour réassigner une console).
2.	Une session ↔ au plus un terminal
•	Une session n’a qu’un seul ctty possible.
•	C’est lui qui distribue les signaux clavier (SIGINT, SIGTSTP, …) au foreground process group de la session.

⸻

🔹 Exemple concret

/dev/pts/3
|
└── Session #1 (leader = bash)  ← ctty = /dev/pts/3
|
├── Foreground: vim
└── Background: sleep

	•	Ici, /dev/pts/3 est lié à la session de bash.
	•	Si un autre process lance setsid() et tente TIOCSCTTY sur /dev/pts/3, il sera refusé.

⸻

🔹 Quand plusieurs sessions semblent partager un terminal

Ça peut arriver si :
•	tmux/screen → ils créent un PTY par session interne et multiplexent le vrai terminal en arrière-plan.
•	ssh → chaque connexion SSH a son propre PTY (nouveau ctty).
•	getty/login → au logout/login, ils réassignent la console au nouvel utilisateur (c’est une ré-attribution, pas un partage).

👉 Mais dans le kernel, à un instant donné :
un terminal = au plus un ctty = une seule session.

⸻

✅ Donc :
•	Plusieurs sessions ne peuvent pas posséder le même ctty en même temps.
•	Elles doivent chacune avoir leur propre terminal (vrai TTY ou PTY), ou bien passer par un multiplexeur comme screen/tmux.

⸻

Veux-tu que je t’explique comment tmux/screen font croire à plusieurs sessions interactives sur le même terminal physique ?

----

Exactement ✅

Dans une même session, tous les processus héritent du même ctty (controlling terminal).
Mais attention : ils ne l’utilisent pas tous de la même manière → c’est là qu’entrent en jeu les process groups et le foreground/background.

⸻

🔹 Règles dans une session avec un ctty
•	Tous les process de la session connaissent le ctty
(ils peuvent l’ouvrir via /dev/tty ou /dev/pts/N).
•	Un seul process group est au foreground
•	C’est celui qui reçoit les signaux clavier (Ctrl-C = SIGINT, Ctrl-Z = SIGTSTP, Ctrl-\ = SIGQUIT, SIGWINCH au resize).
•	Les autres process groups (background) n’ont pas ces signaux.
•	E/S protégées par signaux
•	Si un process en background tente de lire depuis le ctty → il reçoit SIGTTIN.
•	S’il tente d’écrire → il reçoit SIGTTOU.
•	Ça évite que des jobs en arrière-plan polluent ton terminal interactif.

⸻

🔹 Exemple concret

/dev/pts/3 (ctty)
Session leader: bash
|
+--> Process group #1 (foreground)
|        └── [vim]
|
+--> Process group #2 (background)
└── [sleep 1000]

	•	Les deux (vim, sleep) partagent le même ctty (ils sont dans la même session).
	•	Mais seul vim (foreground) reçoit SIGINT si tu tapes Ctrl-C.
	•	Si sleep essaie de lire depuis le terminal → kernel lui envoie SIGTTIN.

⸻

🔹 Résumé
•	Oui → dans une session, plusieurs processus partagent le même ctty.
•	Mais → seul le process group au foreground a le droit de l’utiliser normalement.
•	Les autres process groups sont limités par le kernel via SIGTTIN/SIGTTOU.

⸻

👉 Veux-tu que je te montre un exemple en bash où tu vois concrètement un job background qui reçoit SIGTTIN parce qu’il essaie de lire depuis le terminal ?