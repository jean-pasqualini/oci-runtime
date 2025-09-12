exactement 👌 tu as la bonne intuition :

séquence réelle côté runc
1.	Préparer la commande runc init

cmd := exec.Command("/proc/self/exe", "init", "--some-flags")

(en gros runc s’auto-re-exécute en mode init).

	2.	cmd.Start()
	•	ça fork/exec l’init.
	•	tout de suite après, le parent (runc create ou runc run) a les deux extrémités des pipes (sync/log).
	3.	Communication avec l’init
	•	le parent écrit le gros blob JSON initConfig dans le pipe de sync → l’init lit et se configure.
	•	ensuite, petit protocole de messages (procReady, procRun, …) sur ce pipe.
	•	si ça échoue, l’init écrit l’erreur sur le log pipe, que le parent affiche.
	4.	Selon la commande (create ou run) :
	•	runc create :
	•	fait le cmd.Start() et tout le handshake,
	•	laisse l’init bloqué sur exec.fifo,
	•	puis ne fait pas de cmd.Wait(), il quitte en laissant l’init tourner → conteneur "created".
	•	runc run :
	•	même cmd.Start() + handshake,
	•	débloque exec.fifo pour que l’init exécute ton process,
	•	puis appelle cmd.Wait() pour rester attaché et se terminer quand ton process se termine.

⸻

en pseudo-Go ça donne :

// commun aux deux
cmd := exec.Command("/proc/self/exe", "init")
cmd.Start()                 // lance l’init
sendInitConfigJSON(cmd)     // gros blob sur le pipe
handshakeProcReady(cmd)     // procReady/procRun etc.

// différence
if mode == "create" {
// on sort, init reste bloqué → état created
return
}
if mode == "run" {
unblockExecFifo()       // débloque init pour execve du vrai process
cmd.Wait()              // attendre la fin du process container
}


⸻

👉 donc oui : en pratique, même dans runc run, il y a bien un Start() au début, suivi du handshake JSON/pipes, et le Wait() vient seulement après que tout est configuré et démarré.

Veux-tu que je te fasse un petit schéma timeline (parent vs init) montrant où tombent Start(), initConfig, procReady/procRun, exec.fifo, et Wait() ?