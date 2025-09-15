Bonne idée 👍
Les histoires de TTY/PTY, setsid, TIOCSCTTY sont pleines de pièges. Voilà un cheatsheet synthétique pour éviter de te refaire avoir.

⸻

📝 Cheatsheet TTY / PTY / setsid / TIOCSCTTY

Concepts de base
•	TTY : terminal “classique” (physique ou émulateur).
•	PTY : pseudo-terminal (paire master/slave).
•	master côté programme qui pilote (ex: ton runtime).
•	slave côté programme à exécuter (ex: /bin/sh dans le conteneur).
•	Controlling terminal : un terminal associé à une session → reçoit signaux (SIGHUP, SIGINT, SIGTSTP, …).
•	Process group / session :
•	setsid() crée une nouvelle session + groupe de process et détache du TTY précédent.

⸻

Recettes pratiques

1. Créer un PTY

#include <pty.h>   // ou openpty()
int master, slave;
openpty(&master, &slave, NULL, NULL, NULL);

	•	master reste côté runtime.
	•	slave doit être donné au process dans le conteneur.

⸻

2. Isoler ton process (dans l’init du conteneur)

setsid();                                // nouvelle session
ioctl(slave, TIOCSCTTY, 0);              // devient le TTY de contrôle
dup2(slave, STDIN_FILENO);
dup2(slave, STDOUT_FILENO);
dup2(slave, STDERR_FILENO);
close(slave);                            // fd plus nécessaire
execve("/bin/sh", argv, envp);

✅ Le shell croit parler à un vrai terminal.
⚠️ À ne jamais faire sur le TTY hérité du parent (sinon tu lui “voles” son terminal).

⸻

3. Côté runtime
   •	Proxifier entre master et tes stdio :

// lire/écrire master <-> stdin/stdout


	•	Gérer la taille du terminal :

struct winsize ws;
ioctl(STDIN_FILENO, TIOCGWINSZ, &ws);   // lire taille host
ioctl(master, TIOCSWINSZ, &ws);         // appliquer au pty


	•	Propager SIGWINCH pour redimensionner.

⸻

4. Signaux liés au TTY
   •	SIGHUP : envoyé si le TTY de contrôle disparaît (ou si le leader de session ferme son TTY).
   •	SIGINT (Ctrl+C), SIGTSTP (Ctrl+Z) : envoyés au foreground process group du TTY.
   •	SIGCONT : reprise d’un job stoppé.

⸻

5. Cas fréquents / Pièges
   •	Si tu fais TIOCSCTTY sur le TTY du bash parent ⇒ bash perd son TTY ⇒ SIGHUP ⇒ bash se ferme.
   •	Si tu fermes le dernier fd vers le TTY du parent ⇒ même punition.
   •	Si tu ne fais pas setsid() avant TIOCSCTTY ⇒ EPERM.
   •	Sans TIOCSCTTY ⇒ ton shell tourne mais ne réagit pas comme interactif (pas de job control, affichage bizarre).
   •	Sans -i (ou terminal ouvert), /bin/sh quitte dès que stdin est EOF.

⸻

Résumé “1-liner”

👉 Toujours :
setsid(); ioctl(slave_fd, TIOCSCTTY, 0); dup2(slave_fd,0/1/2); execve(...)
sur un PTY neuf créé par le runtime.
❌ Jamais sur le TTY hérité du parent.

⸻

Veux-tu que je te fasse aussi une version Go (avec github.com/creack/pty) de ce pattern pour que tu l’intègres directement dans ton runtime maison ?

-----

Non 🙂

Quand tu lances un programme “normal” depuis ton shell (bash, zsh…), il ne vole pas le pty.
Voici ce qui se passe :

⸻

Cas standard (shell → programme)
•	Ton shell tourne dans une session avec un controlling TTY (ex: /dev/pts/3).
•	Quand tu fais ls ou vim, ton shell fait un fork() puis execve("ls", …).
•	Le processus enfant hérite des FDs 0/1/2 → donc il utilise le même pty que le shell.
•	Pas besoin de refaire setsid() ni TIOCSCTTY, puisque la session et le controlling TTY existent déjà (ceux du shell).
•	Quand le programme quitte, le shell reprend la main sur les mêmes FDs → tout continue normalement.

Donc : pas de vol → le programme n’a rien changé à la session ni au terminal.

⸻

Cas où ça peut “voler”

Le seul moment où un process peut voler un pty, c’est s’il appelle volontairement :
•	setsid() pour devenir leader d’une nouvelle session et
•	ioctl(fd, TIOCSCTTY) sur un pty/tty qu’il a ouvert.

👉 Ça n’arrive pas quand tu fais ls depuis bash. Mais ça peut arriver avec des programmes comme login, agetty, ssh, ou ton runtime si tu le codes ainsi.

⸻

Résumé rapide
•	Programme lancé depuis le shell → hérite du pty, ne le vole pas.
•	Programme qui veut devenir “leader de session avec son propre TTY” → doit faire setsid()+TIOCSCTTY, et là il prend le TTY comme controlling terminal → ce qui peut “voler” le TTY au shell si c’est le même FD.

⸻

👉 Donc ce que tu as observé dans ton runtime, c’est probablement que ton init a refait cette séquence (setsid()+TIOCSCTTY) sur le pty du bash → d’où le SIGHUP sur ton shell.

⸻

Tu veux que je te trace en détail ce qui se passe avec un strace quand tu fais par exemple ls vs quand tu fais login/agetty (qui font, eux, un TIOCSCTTY) ?

-----


Exactement ✅

⸻

1. Processus “normal” lancé par le shell
   •	Ton shell est déjà leader de session (il a fait setsid() quelque part dans son histoire — souvent lancé par login ou systemd).
   •	Les programmes que tu lances (ls, vim, etc.) héritent de :
   •	la même session,
   •	le même groupe de processus (sauf si bash les met dans un autre PG pour le job control),
   •	le même controlling TTY (ton /dev/pts/N).
   •	Ils n’ont rien besoin de faire → ils empruntent simplement le TTY.
   •	À la fin, ils sortent, le shell récupère la main.

⸻

2. Devenir leader de session

La commande système pour ça, c’est l’appel système :

setsid();

	•	Crée une nouvelle session dont ton process devient leader.
	•	Détache ton process de tout controlling TTY existant.
	•	Ton process ne peut plus accéder au TTY de la session précédente comme terminal de contrôle.

⸻

3. S’approprier un TTY

Une fois leader de session, tu peux appeler :

ioctl(fd, TIOCSCTTY, 0);

sur un descripteur ouvert vers un terminal (ex : un PTY slave).
•	Ça fait de ce terminal le controlling TTY de ta nouvelle session.
•	À partir de là, ton process et ses descendants reçoivent les signaux du terminal (SIGINT, SIGHUP, etc.).

⸻

4. Le “vol”
   •	Si tu fais setsid() + TIOCSCTTY sur le TTY de quelqu’un d’autre (par ex. le /dev/pts/N du bash), tu lui retires son terminal de contrôle → il prend un SIGHUP → il se ferme.
   •	C’est ce qui t’est arrivé dans ton runtime maison.

⸻

🔑 Règle simple
•	Pas de setsid() = tu restes dans la session existante → tu empruntes le TTY.
•	Avec setsid() + TIOCSCTTY = tu deviens leader d’une nouvelle session et tu t’appropries le TTY.

⸻

Tu veux que je t’écrive un mini exemple C qui montre :
1.	un process qui hérite du TTY (simple execve)
2.	un process qui fait setsid()+TIOCSCTTY et “vole” le terminal ?


----

Très bonne question 👍
Les process groups (groupes de processus) et les sessions sont la brique de base pour comprendre comment un terminal distribue ses signaux (Ctrl+C, Ctrl+Z, SIGHUP, etc.).

⸻

🧩 Les notions

1. Process group (PGID)
   •	Un process group est un ensemble de processus identifiés par un PGID.
   •	Un process group a toujours un leader (le process dont le PID = PGID).
   •	Sert à gérer des ensembles de processus ensemble (ex. une pipeline ls | grep → tous les deux dans le même process group).

2. Session (SID)
   •	Une session est un ensemble de process groups.
   •	Le premier process qui appelle setsid() devient leader de session (SID = son PID).
   •	Une session peut avoir un controlling TTY → un terminal associé.

3. Foreground vs Background
   •	Le terminal a un foreground process group : c’est celui qui reçoit :
   •	l’input (read sur /dev/pts/N)
   •	les signaux générés par le terminal : SIGINT (Ctrl+C), SIGTSTP (Ctrl+Z), SIGQUIT (Ctrl+)
   •	Les autres process groups de la session sont en background :
   •	s’ils essaient de lire sur le terminal → SIGTTIN
   •	s’ils essaient d’écrire → SIGTTOU

⸻

🔨 Exemple concret

Depuis bash

$ sleep 100 | cat

	•	bash crée une pipeline.
	•	Les deux processus (sleep, cat) sont mis dans le même process group.
	•	Ce process group est marqué comme foreground sur ton /dev/pts/N.

Si tu tapes Ctrl+C :
•	Le noyau envoie SIGINT à tout le process group → sleep et cat meurent ensemble.

Avec &

$ sleep 100 &

	•	Ici sleep est dans un autre process group → marqué background.
	•	Le shell garde son propre PG comme foreground.
	•	Si tu fais Ctrl+C, c’est seulement bash (et ses foreground jobs) qui prennent le signal, pas sleep.

⸻

🖇️ En lien avec ton runtime
•	Quand tu lances un conteneur, si ton init reste dans le PG du bash, il “participe” au job-control du shell.
•	Si ton init fait setsid() :
•	Il devient leader d’une nouvelle session.
•	Il sort du process group du bash → isolation.
•	Ensuite, il peut s’attacher à son propre PTY avec TIOCSCTTY.
•	C’est exactement ce que font login, sshd, runc, etc. pour donner un terminal isolé à ton process PID1.

⸻

TL;DR
•	Process group = groupe logique de process pour les signaux et job-control.
•	Foreground PG du TTY = seul groupe qui reçoit Ctrl+C, Ctrl+Z.
•	Session = ensemble de process groups, éventuellement associé à un TTY.
•	setsid() crée une nouvelle session et un nouveau process group → ton process devient isolé.
•	TIOCSCTTY attache un TTY à cette nouvelle session.

⸻

👉 Veux-tu que je te fasse un petit schéma ASCII montrant “bash + PGID”, puis “pipeline”, puis “nouvelle session avec setsid()” ?

----

ps -o pid,ppid,pgid,sid,tty,cmd

----

Tu tiens le bon bout ✅ mais les rôles sont légèrement différents. On peut résumer comme ça :

⸻

📦 Process Group (PGID)
•	Sert à grouper des processus pour le job control.
•	C’est au niveau du PG que le noyau envoie les signaux liés au terminal :
•	SIGINT (Ctrl+C), SIGTSTP (Ctrl+Z), SIGQUIT (Ctrl+\) → envoyés au foreground process group du TTY.
•	En background, si un process groupe tente de lire depuis le TTY → SIGTTIN.
•	S’il tente d’écrire → SIGTTOU.

👉 Les process groups = unité de signalisation pour le job-control.

⸻

📡 Session (SID)
•	Une session regroupe un ou plusieurs process groups.
•	Une session peut avoir un controlling terminal (via setsid() + TIOCSCTTY).
•	Le leader de session (SID = PID du leader) est celui qui a “réclamé” le TTY comme terminal de contrôle.
•	Le TTY est donc associé à la session entière, pas à un seul process group.

👉 Les sessions = unité d’attachement au TTY (controlling terminal).

⸻

🔗 Relation entre les deux
•	SID → “qui possède le TTY” (le leader de session qui a fait TIOCSCTTY).
•	PGID → “qui reçoit les signaux du TTY” (le foreground PGID de ce TTY).

Exemple classique :

bash (SID=100, PGID=100)
└─ sleep 100 (PGID=200)

	•	Les deux sont dans la même session (SID=100) → même controlling TTY.
	•	sleep est dans un autre process group (PGID=200).
	•	Si bash met PGID=200 en foreground (tcsetpgrp), alors Ctrl+C envoie SIGINT au groupe 200 (donc sleep meurt).
	•	Si PGID=100 reste foreground, alors Ctrl+C touche bash, pas sleep.

⸻

📝 TL;DR
•	Sessions (SID) : gèrent l’association avec le TTY (qui a le droit de l’avoir comme terminal de contrôle).
•	Process groups (PGID) : gèrent la distribution des signaux liés à ce TTY (foreground vs background).

⸻

Veux-tu que je te fasse un petit schéma ASCII “TTY ↔ Session ↔ Process groups” pour bien visualiser qui possède le TTY et qui reçoit les signaux ?

--------

Parfait 👌, on va plonger dans le foreground / background des process groups : c’est le cœur du job control des shells (et donc ce qui t’embête avec ton runtime maison).

⸻

🖇️ Rappel : sessions, PG, TTY
•	Session : ensemble de process groups, éventuellement attachée à un controlling terminal.
•	Foreground process group : celui qui a le contrôle du TTY → reçoit l’input et les signaux (Ctrl+C, Ctrl+Z…).
•	Background process groups : les autres groupes de la session → n’ont pas le droit de lire/écrire librement sur le TTY.

⸻

⚙️ Foreground vs Background en pratique

1. Foreground PG
   •	Le noyau maintient pour chaque TTY un PGID foreground (tcgetpgrp() / tcsetpgrp()).
   •	Ce PGID reçoit :
   •	les signaux de terminal :
   •	SIGINT (Ctrl+C)
   •	SIGQUIT (Ctrl+\)
   •	SIGTSTP (Ctrl+Z)
   •	les données lues depuis le clavier (read()).
   •	Les process de ce groupe peuvent écrire directement sur le terminal.

2. Background PG
   •	Tout autre process group de la session est background.
   •	Si un process en background :
   •	essaie de lire → noyau envoie SIGTTIN au groupe (par défaut, il s’arrête).
   •	essaie d’écrire (si tostop est activé dans stty) → noyau envoie SIGTTOU.

⸻

🔨 Exemple concret avec bash

$ sleep 100 | cat

	•	sleep et cat dans le même PGID (mettons 1234).
	•	bash dans PGID 5678.
	•	Bash fait tcsetpgrp(tty, 1234) → groupe 1234 devient foreground.
	•	Ctrl+C → noyau envoie SIGINT à tout le PG 1234 → sleep et cat meurent.

⸻


$ sleep 100 &

	•	sleep dans PGID 1235, marqué background.
	•	bash garde son PGID 5678 comme foreground.
	•	Si sleep tente de lire au clavier (read()) → SIGTTIN, il se suspend.
	•	Ctrl+C → signalé au PG 5678 (donc bash), pas à sleep.

⸻

🛠️ Les appels systèmes clés
•	tcgetpgrp(fd) : lire le PGID au foreground du TTY.
•	tcsetpgrp(fd, pgid) : définir un PG comme foreground du TTY.
(C’est ce que fait bash quand tu tapes fg %1).

⸻

🚦 TL;DR
•	Foreground PGID = qui reçoit l’input + signaux (Ctrl+C, Ctrl+Z).
•	Background PGID = dans la session mais pas au contrôle du TTY → bloqué ou stoppé s’il tente d’interagir.
•	Le shell manipule ça en appelant setpgid() (pour créer des PG) puis tcsetpgrp() (pour basculer foreground/background).

⸻

👉 Veux-tu que je te montre un petit pas-à-pas reproductible (sleep, cat, jobs, fg, bg, ps -o pid,pgid,sid) pour voir le foreground PG basculer en direct ?

----


Parfait 👌, on va plonger dans le foreground / background des process groups : c’est le cœur du job control des shells (et donc ce qui t’embête avec ton runtime maison).

⸻

🖇️ Rappel : sessions, PG, TTY
•	Session : ensemble de process groups, éventuellement attachée à un controlling terminal.
•	Foreground process group : celui qui a le contrôle du TTY → reçoit l’input et les signaux (Ctrl+C, Ctrl+Z…).
•	Background process groups : les autres groupes de la session → n’ont pas le droit de lire/écrire librement sur le TTY.

⸻

⚙️ Foreground vs Background en pratique

1. Foreground PG
   •	Le noyau maintient pour chaque TTY un PGID foreground (tcgetpgrp() / tcsetpgrp()).
   •	Ce PGID reçoit :
   •	les signaux de terminal :
   •	SIGINT (Ctrl+C)
   •	SIGQUIT (Ctrl+\)
   •	SIGTSTP (Ctrl+Z)
   •	les données lues depuis le clavier (read()).
   •	Les process de ce groupe peuvent écrire directement sur le terminal.

2. Background PG
   •	Tout autre process group de la session est background.
   •	Si un process en background :
   •	essaie de lire → noyau envoie SIGTTIN au groupe (par défaut, il s’arrête).
   •	essaie d’écrire (si tostop est activé dans stty) → noyau envoie SIGTTOU.

⸻

🔨 Exemple concret avec bash

$ sleep 100 | cat

	•	sleep et cat dans le même PGID (mettons 1234).
	•	bash dans PGID 5678.
	•	Bash fait tcsetpgrp(tty, 1234) → groupe 1234 devient foreground.
	•	Ctrl+C → noyau envoie SIGINT à tout le PG 1234 → sleep et cat meurent.

⸻


$ sleep 100 &

	•	sleep dans PGID 1235, marqué background.
	•	bash garde son PGID 5678 comme foreground.
	•	Si sleep tente de lire au clavier (read()) → SIGTTIN, il se suspend.
	•	Ctrl+C → signalé au PG 5678 (donc bash), pas à sleep.

⸻

🛠️ Les appels systèmes clés
•	tcgetpgrp(fd) : lire le PGID au foreground du TTY.
•	tcsetpgrp(fd, pgid) : définir un PG comme foreground du TTY.
(C’est ce que fait bash quand tu tapes fg %1).

⸻

🚦 TL;DR
•	Foreground PGID = qui reçoit l’input + signaux (Ctrl+C, Ctrl+Z).
•	Background PGID = dans la session mais pas au contrôle du TTY → bloqué ou stoppé s’il tente d’interagir.
•	Le shell manipule ça en appelant setpgid() (pour créer des PG) puis tcsetpgrp() (pour basculer foreground/background).

⸻

👉 Veux-tu que je te montre un petit pas-à-pas reproductible (sleep, cat, jobs, fg, bg, ps -o pid,pgid,sid) pour voir le foreground PG basculer en direct ?