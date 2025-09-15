Très bonne question 👌 ça touche au cœur de la différence entre “hériter du TTY du parent” et “obtenir un nouveau PTY master/slave” via un console socket.

⸻

🔹 1. TTY hérité
•	Ton shell (ou ton parent) a déjà son stdin/stdout/stderr branché sur un TTY (ex : /dev/pts/0).
•	Quand il fork/exec un programme, les FDs 0/1/2 sont hérités tels quels.
•	Le process enfant est donc attaché au même TTY que son parent.
•	Exemple typique :

bash
└─ runc run myctr
└─ /bin/sh   (dans le conteneur, hérite du même /dev/pts/0)


	•	Ici, pas besoin de console socket : runc transmet juste les FDs.

⸻

🔹 2. Nouveau PTY master/slave via console socket
•	runc peut créer un nouveau terminal dédié au conteneur.
•	Pour ça, il a besoin d’un couple PTY master/slave :
•	le slave est donné comme stdin/stdout/stderr du process dans le conteneur,
•	le master doit être géré par quelqu’un à l’extérieur (orchestrateur).
•	Comme runc ne garde pas de “shim” permanent, il ne veut pas lui-même gérer le master.
•	Donc : il demande à un console socket (un autre process comme containerd-shim, systemd, ou un helper) de créer le PTY master, et de lui passer le slave par un passage de descripteur (SCM_RIGHTS sur un socket Unix).

Exemple avec runc run --tty --console-socket /tmp/console.sock :
•	runc : “j’ai besoin d’un terminal”
•	l’orchestrateur (via le socket) : crée le master/slave, garde le master pour gérer l’I/O, envoie le slave à runc
•	runc : branche le slave comme stdin/stdout/stderr du conteneur.
•	L’orchestrateur peut ensuite attacher/détacher, redimensionner la fenêtre (SIGWINCH), etc.

⸻

🔹 Différence clé
•	TTY hérité
•	Avantage : simple, direct, zéro infra.
•	Limite : le conteneur partage le terminal du parent, tu ne peux pas détacher facilement, ni donner un terminal dédié.
•	TTY créé via master/slave et console socket
•	Avantage : terminal isolé, redirigeable, multiplexable.
•	Permet l’attache/détache (docker attach style), resize de fenêtre, logging, etc.
•	C’est ce mécanisme qu’utilisent containerd, docker, cri-o → ils gardent le master côté hôte et passent le slave au conteneur.

⸻

⚡ En résumé
•	Hérité → ton conteneur utilise le même TTY que toi.
•	Master/slave via console socket → ton conteneur a son propre TTY dédié, relié au monde extérieur par un orchestrateur.

⸻

👉 Veux-tu que je te fasse un petit schéma ASCII pour visualiser clairement les deux cas (hérité vs master/slave via console socket) ?