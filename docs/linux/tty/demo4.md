Très bonne question 👌 tu mets le doigt sur deux mécanismes différents de gestion de vie et de mort des processus :

⸻

1. Pourquoi ^C (SIGINT) n’arrête pas les jobs en arrière-plan ?
   •	Quand tu tapes ^C :
   •	Le TTY envoie SIGINT au foreground process group uniquement.
   •	Les background process groups (autres PGID rattachés au même TTY) ne reçoivent rien.
   •	Résultat : seul ton job au premier plan meurt, les jobs en arrière-plan continuent.

👉 C’est exactement pour ça qu’on peut lancer :

sleep 100 &

puis faire ^C → ton shell reste, et le sleep en arrière-plan n’est pas touché.

⸻

2. Pourquoi si le shell meurt, les jobs meurent aussi ?

Ici, ce n’est pas SIGINT qui joue, mais la gestion des sessions et TTY par le noyau :
•	Ton bash est le session leader (STAT s).
•	Il détient le TTY de contrôle (/dev/pts/0).
•	Quand le shell (session leader) meurt :
•	Le noyau détache le TTY de tous les autres processus encore attachés.
•	Le noyau leur envoie SIGHUP (hangup), pour dire : “votre terminal de contrôle a disparu”.
•	Par défaut, beaucoup de programmes quittent quand ils reçoivent SIGHUP.

⸻

3. Que se passe-t-il si on ne veut pas que les jobs meurent ?

C’est pour ça qu’existent :
•	nohup → ignore SIGHUP :

nohup long_process &


	•	disown (bash/zsh) → détache un job du shell, il ne recevra pas SIGHUP si le shell meurt :

long_process &
disown %1


	•	multiplexeurs comme screen / tmux → le TTY “virtuel” reste vivant même si ton shell parent se ferme.

⸻

✅ Résumé
•	^C = SIGINT → seulement au foreground process group → jobs background intacts.
•	Shell meurt = SIGHUP du noyau → tous les jobs de la session reçoivent le signal → ils meurent (sauf protégés par nohup, disown, tmux, etc.).

⸻

👉 Tu veux que je te fasse un petit tableau comparatif (signal, qui le reçoit, effet sur foreground/background) pour avoir une vue claire de ^C, ^Z, exit, kill, logout, etc. ?