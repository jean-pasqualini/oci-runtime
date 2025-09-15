ps -A -o user,tty,pgid,stat,sid,command --forest | grep -E "USER|pts/0" | grep -v "grep"
watch -n 0.1 'ps -A -o user,tty,pgid,stat,sid,command --forest | grep -E "USER|pts/0" | grep -v "grep"'

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