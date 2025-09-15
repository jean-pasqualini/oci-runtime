
Yes 👍 tu as mis le doigt sur une vraie subtilité : savoir si ton conteneur meurt parce que le shell n’a pas de TTY ou parce qu’il lit sur un stdin inexistant.

⸻

🔹 Le problème avec sh / bash
•	Quand /bin/sh ou /bin/bash démarre, ils testent si stdin est un tty.
•	S’il n’y a pas de TTY (ou pas de stdin lisible), le shell peut se fermer immédiatement (surtout s’il n’a rien à exécuter).
•	Donc :
•	parfois ton conteneur s’arrête “trop vite” → pas clair si c’est un bug du TTY ou juste le shell qui sort.

⸻

🔹 Solutions : binaires simples pour tester

Tu peux mettre dans ton config.json un processus minimal qui ne dépend pas du TTY :

1. cat

"process": { "args": ["cat"] }

	•	cat lit son stdin et recopie vers stdout.
	•	Si pas de TTY → il lit sur un pipe.
	•	Tu verras si le process reste bloqué (lit) ou s’arrête.

2. yes

"process": { "args": ["yes"] }

	•	Écrit “y” en boucle sur stdout.
	•	Ne lit jamais stdin → indépendant du TTY.
	•	S’arrête uniquement si tu le tues.

3. sleep

"process": { "args": ["sleep", "1000"] }

	•	Ne touche jamais au TTY.
	•	Utile pour voir si le conteneur reste vivant.

4. echo

"process": { "args": ["echo", "hello"] }

	•	Écrit et sort immédiatement.
	•	Permet de tester juste l’attachement de stdout.

⸻

🔹 Pourquoi c’est utile
•	Si sleep 1000 reste vivant → ton problème vient bien du shell qui sort faute de TTY.
•	Si même sleep 1000 sort immédiatement → ton problème est ailleurs (ex : runc start mal configuré, rootfs manquant, etc.).

⸻

🔹 Bonus

Tu peux aussi écrire toi-même sur le tty à la main, par exemple avec :

echo "coucou" > /dev/tty

Mais ça ne marche que si /dev/tty existe dans le conteneur (lié à ton stdin).
Avec yes ou sleep, tu t’affranchis complètement de ça.

⸻

👉 Tu veux que je te prépare un config.json minimal qui remplace /bin/sh par sleep 1000 pour que tu puisses tester direct si ton conteneur crashe à cause du TTY ?

----


Yes, c’est exactement le sujet des process groups et du foreground.
Même si tu as un vrai TTY, si ton process n’est pas dans le process group au foreground du TTY, une lecture dessus recevra SIGTTIN (et sera stoppée), et une écriture peut recevoir SIGTTOU (si tostop est actif).

Ce qui se passe
•	Lecture depuis le TTY quand on n’est pas au foreground → SIGTTIN → le process passe en état T (stopped).
•	Écriture depuis l’arrière-plan → SIGTTOU si stty tostop (par défaut souvent -tostop, donc écriture autorisée).

Comment vérifier vite (depuis le conteneur)
1.	Le descripteur est-il un TTY ?

[ -t 0 ] && echo "stdin est un TTY" || echo "stdin n'est PAS un TTY"

	2.	Suis-je au foreground de CE TTY ?

	•	Méthode simple (bash) : le + dans STAT = foreground

ps -o stat= -p $$    # contient un '+' si ce shell est au foreground

	•	Méthode précise (compare pgrp vs fg pgrp du TTY) :

# nécessite Python (ou perl)
python3 - <<'PY'
import os, sys
fg = os.tcgetpgrp(sys.stdin.fileno())   # pgrp au foreground du TTY
me = os.getpgrp()                       # pgrp de mon process
print("FG" if fg == me else f"BG (fg_pgrp={fg}, my_pgrp={me})")
PY

Si tu vois BG (...), alors ton process n’est pas au foreground ⇒ une lecture bloquera (SIGTTIN).

Reproduire / diagnostiquer
•	Lancer un lecteur bloquant :

cat    # ou 'read line' dans bash

S’il se met en stopped sans rien afficher, c’est SIGTTIN (pas au foreground).
Vérifie avec jobs/ps -o stat= → T (stopped).

Que faire côté runc
•	runc run --tty --console-socket ... et “attach” : l’orchestrateur qui tient le master PTY doit faire un tcsetpgrp() pour mettre le pgrp du conteneur au foreground quand tu es attaché. C’est ce que font containerd-shim/docker en pratique.
•	Si tu fais juste runc start détaché et que ton init (ex: bash) tente de lire le TTY sans être foreground, il se fera stopper par SIGTTIN.
•	Workarounds immédiats si tu ne peux pas mettre le process au foreground :
•	Ne lis pas le TTY : lance une commande qui n’attend pas stdin (sleep 9999, yes, un serveur, etc.).
•	Redirect stdin ailleurs :
args: ["bash","-lc","commande"] avec < /dev/null ou un pipe.
•	Pour écrire sans être stoppé : stty -tostop sur ce TTY (mais ça ne règle pas la lecture).

TL;DR
•	Tu peux avoir un TTY valide ([ -t 0 ] vrai) mais pas être au foreground → les lectures se font stopper (SIGTTIN) ; tty peut même dire not a tty si le device n’a pas de nom dans /dev.
•	Vérifie pgrp vs tcgetpgrp et, côté runtime, assure-toi que l’attach met bien le pgrp du conteneur au foreground (via le console socket qui gère le PTY master).

---


Oui (en partie) — ça dépend si ton process écrit ou lit sur le TTY en étant background.

Écriture depuis l’arrière-plan (SIGTTOU)

Par défaut, l’écriture n’arrête pas un job background (option -tostop désactivée).
Ça ne “tue”/stoppe que si tostop est activé.
•	Vérifier :

stty -a | grep tostop
# ... -tostop ...   => OK (n’arrête pas)
# ... tostop ...    => Mauvais (arrête à l’écriture)


	•	Désactiver l’arrêt à l’écriture :

stty -tostop


	•	Ou ignorer le signal côté process (empêche l’arrêt même si tostop est actif) :

trap '' TTOU   # en shell
# ou en C: signal(SIGTTOU, SIG_IGN);



Lecture depuis l’arrière-plan (SIGTTIN)

Là, tu ne peux pas vraiment “désactiver” le comportement noyau : si un process pas au foreground lit le TTY, il reçoit SIGTTIN (et est stoppé).

Pour éviter l’arrêt :
1.	Mettre le process au foreground du TTY
→ l’entité qui tient le master pty doit faire tcsetpgrp() vers le pgrp du process (c’est ce que fait un “attach” bien implémenté).
2.	Ne pas lire le TTY
•	Rediriger l’entrée :

cmd < /dev/null

(ou un pipe/fichier)

	•	Choisir un binaire qui n’attend pas stdin (ex. sleep, yes n’a pas de lecture).

	3.	Ignorer SIGTTIN
	•	trap '' TTIN (ou signal(SIGTTIN, SIG_IGN)).
	•	Effet pratique : la lecture échoue (souvent read() -> -1 EIO), mais le process n’est plus stoppé.
	4.	Supprimer la notion de TTY
	•	Démarrer dans une nouvelle session sans terminal :

setsid cmd </dev/null >/dev/null 2>&1 &


	•	Dans runc, mettre "terminal": false (pas de TTY), ou lancer détaché et rediriger stdin.

Récap express
•	Écriture: désactive tostop (stty -tostop) ou ignore SIGTTOU.
•	Lecture: pas de “désactivation” magique ; soit tu deviens foreground (tcsetpgrp via l’orchestrateur/attach), soit tu n’essaies pas de lire, soit tu ignores SIGTTIN (la lecture échoue mais le process continue), soit tu retires le TTY.

Si tu me dis ton cas précis (process lit/écrit ? tu as --tty ? attaché ou non ?), je te donne la ligne exacte à mettre dans ton config.json ou la commande runc adaptée.


----

Très bonne question 👌. Ça touche à la mécanique du reparenting et des TTY sous Linux.

⸻

🔹 Quand un process est reparenté

Si ton process est adopté (par ex. runc create → runc:[1:INIT], ou par init/systemd si le parent meurt) :
•	Ses file descriptors ouverts ne changent pas.
•	Donc si /proc/<pid>/fd/0 → /dev/pts/0, ça reste pareil : FD0 est toujours ce PTY.
•	Le reparenting ne casse pas la liaison au TTY.

⸻

🔹 Mais ce qui peut changer

Le vrai sujet, c’est la relation process group ↔ foreground process group du TTY :
•	Le PTY /dev/pts/0 a un foreground process group (tcgetpgrp(fd)).
•	Ton process a son propre process group (getpgrp()).
•	Tant que getpgrp() == tcgetpgrp(fd) → il est au foreground → il peut lire le TTY.
•	Si ≠ → il est en background :
•	lecture → SIGTTIN → process stoppé,
•	écriture → possible, sauf si stty tostop.

Quand le parent meurt et que ton process est reparenté, deux cas :
1.	Il garde son pgrp → pas de problème a priori.
2.	L’orchestrateur ou le nouveau parent ne fait pas de tcsetpgrp sur le PTY → ton process se retrouve “background” vis-à-vis de son propre TTY. Résultat :
•	/proc/self/fd/0 pointe toujours /dev/pts/0
•	mais lire dessus = SIGTTIN (stoppé)

⸻

🔹 Exemple concret

Imaginons un conteneur lancé avec TTY :

ls -l /proc/self/fd/0
# → /dev/pts/0

	•	Si runc run --tty (attaché direct) → tcsetpgrp est fait → ton process est foreground → bash marche.
	•	Si runc create/start détaché (pas d’attach) → le PTY existe toujours (/dev/pts/0), FD0 est là, mais le foreground pgrp n’est pas ton process group → dès que bash lit → SIGTTIN. Tu vois ton process stoppé (ps → STAT=T).

⸻

🔹 Donc
•	Le reparenting ne ferme pas /dev/pts/0.
•	Mais sans orchestration (tcsetpgrp), ton process peut perdre le foreground et se faire bloquer.
•	C’est exactement pourquoi Docker/containerd ont un shim : il garde le PTY master, fait tcsetpgrp, gère les signaux (SIGWINCH, SIGINT, etc.), et assure que ton process dans le conteneur croit être au foreground.

⸻

👉 Veux-tu que je te montre un mini code C (10 lignes) que tu peux lancer dans ton conteneur pour imprimer :
•	isatty(0)
•	getpgrp()
•	tcgetpgrp(0)
et voir si ton process est bien au foreground ou pas ?