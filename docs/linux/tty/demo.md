https://www.linusakesson.net/programming/tty/

TTY, SID, PGID
FOREGROUND PGID
BACKGROUND PGID

ps -A -o user,tty,pgid,sid,stat,command --forest | grep -E "USER|pts/0" | grep -v "grep"

Super 👍 je vais te donner un pas-à-pas que tu peux exécuter dans ton terminal pour voir les changements de process group et de foreground en action.

⸻

🧪 Démo pratique : Foreground / Background PG

1. Prépare la vue avec ps

Ouvre un terminal, et tape :

# on affiche pid, ppid, pgid, sid, tty, cmd
watch -n0.5 'ps -o pid,ppid,pgid,sid,tty,cmd --forest'

👉 Ça rafraîchit toutes les 0,5 secondes, parfait pour observer les changements.

⸻

2. Lance un job au premier plan

Dans un autre terminal (ou un nouveau shell dans le même TTY) :

sleep 100 | cat

	•	Bash crée un nouveau PG pour sleep et cat.
	•	Bash fait tcsetpgrp() pour donner ce PG en foreground.
	•	Dans le watch, tu devrais voir :
	•	sleep et cat → même PGID, différent de bash.
	•	Leur PG est foreground sur ton TTY.

👉 Tape Ctrl+C → les deux process (sleep, cat) meurent car SIGINT est envoyé à tout le PG foreground.

⸻

3. Lance un job en arrière-plan

sleep 1000 &

	•	sleep est mis dans son propre PGID, mais il reste background.
	•	Bash garde son PG en foreground.

👉 Dans le watch, tu verras sleep avec un PGID différent, mais SID = celui de bash.
👉 Si tu tapes Ctrl+C, c’est bash qui reçoit le SIGINT, pas le sleep.

⸻

4. Mets un job au foreground

fg %1

	•	Bash fait un tcsetpgrp(tty, pgid_of_sleep).
	•	Maintenant le PGID du sleep est foreground.
	•	Dans watch, rien ne change visuellement (PGID ne bouge pas), mais le noyau considère ce PG comme foreground.

👉 Tape Ctrl+C → ça tue le sleep.

⸻

5. Mets un job en background

Re-lance sleep 1000, puis fais :

bg %1

	•	Bash garde sleep dans son PG séparé.
	•	Mais il reprend lui-même le foreground (tcsetpgrp(tty, pgid_of_bash)).

👉 sleep continue de tourner, mais si tu fais :

kill -s SIGTTIN <pid_of_sleep>

il se suspendra → car un background PG qui essaie de lire depuis le TTY se fait stopper.

⸻

🔑 Points à observer
•	PGID change quand un nouveau job est créé.
•	SID reste celui du shell (sauf si un process fait setsid()).
•	Foreground PGID n’apparaît pas directement dans ps, mais tu le vois par le comportement aux signaux (Ctrl+C tue soit bash, soit le job).
•	Les transitions sont faites par tcsetpgrp() que bash appelle derrière fg/bg.

⸻

👉 Veux-tu que je te donne aussi la version en C minimaliste qui montre explicitement l’appel à tcgetpgrp / tcsetpgrp pour voir le foreground PGID bouger sans passer par bash ?