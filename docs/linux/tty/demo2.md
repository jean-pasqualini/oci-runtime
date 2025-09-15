Parfait, ton ps est super parlant 👌 Voyons ce que signifient ces colonnes pour déterminer qui est foreground et qui est background.

⸻

📊 Lecture de ton tableau

USER  TT    PGID STAT SID COMMAND
root  pts/0    1 Ss    1 bash
root  pts/0    7 S     1 make run-shim
root  pts/0    7 Sl    1  \_ go run -tags medium ./cmd/container-shim
root  pts/0    7 Sl    1      \_ /tmp/go-build756882639/b001/exe/container-shim
root  pts/0  122 S+    1          \_ /bin/sh -l

Colonnes clés :
•	PGID → identifiant du groupe de processus (jobs).
•	STAT → état du processus :
•	S = sleeping (attente interuptible)

•	D → Uninterruptible sleep (souvent I/O bloquant disque).
•	s = session leader
•	l = multi-threadé
•	T = Stopped (par ^Z ou SIGSTOP).
•	Z = Zombie (fini mais pas encore “reap” par le parent).
•	l = Multi-threadé (CLONE_THREAD).
•	+ = processus dans le foreground group du terminal (c’est le signe que tu cherches).
•	SID → session ID (ton shell est le leader : SID=1).
•	COMMAND → la commande.

⸻

🚀 Analyse foreground / background
•	bash (PGID=1, STAT=Ss) → ton shell interactif, session leader.
•	make run-shim + enfants (PGID=7) → un autre groupe, pas de + → donc ils tournent en arrière-plan.
•	/bin/sh -l (PGID=122, STAT=S+) → marqué + → c’est le foreground process group du TTY pts/0.

⸻

✅ Conclusion
•	+ dans la colonne STAT = foreground job.
•	Les autres PGID (sans +) = background jobs (mais toujours attachés au même terminal).

⸻

👉 Donc dans ton exemple :
•	Foreground job = /bin/sh -l (PGID=122).
•	Background job = make run-shim et toute sa hiérarchie (PGID=7).

⸻

Veux-tu que je te prépare un mini guide visuel des STAT courants (S, R, T, Z, +, etc.) pour que tu puisses interpréter rapidement un ps ?