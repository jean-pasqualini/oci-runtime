Très bonne question 👌

C’est une subtilité du protocole exec.fifo de runc :

⸻

🔹 Rôle du exec.fifo
•	Ce FIFO sert juste de handshake entre les deux parties :
•	runc init (dans le conteneur) → indique « je suis prêt, tu peux me lancer »
•	runc start (côté runtime hôte) → attend ce signal avant de continuer

⸻

🔹 Pourquoi init ouvre en WRONLY
•	Quand init est prêt, il doit notifier start.
•	Le moyen choisi est : ouvrir le FIFO en écriture et y balancer un octet "0".
•	Ce write débloque le open(O_RDONLY) qui a été fait côté runc start.
•	Une fois l’octet écrit, la synchro est faite, start sait qu’il peut lancer le conteneur.

⸻

🔹 Et côté runc start ?
•	C’est start qui ouvre le FIFO en lecture (O_RDONLY).
•	Il bloque jusqu’à ce qu’un writer s’ouvre (l’init).
•	Quand il lit le fameux octet "0", il sait que l’init a fini son setup → il continue.
•	Ensuite il supprime le FIFO (unlink).

⸻

📌 Résumé
•	Init = writer (il envoie le signal).
•	Start = reader (il attend le signal).
•	Un seul octet est suffisant pour débloquer, pas besoin d’un vrai message.

C’est exactement pour ça que dans le code de libcontainer tu vois :

fd, _ := unix.Open(fmt.Sprintf("/proc/self/fd/%d", l.fifoFd), unix.O_WRONLY|unix.O_CLOEXEC, 0)
_, _ = unix.Write(fd, []byte("0"))


⸻

👉 Donc le choix WRONLY côté init est purement lié au protocole de synchro :
c’est lui qui envoie le “go” à runc start.

Veux-tu que je te montre le code côté runc start qui fait le pendant en O_RDONLY ?

----

Bonne question. Techniquement on pourrait inverser, mais runc a choisi Init = writer / Start = reader pour des raisons pratiques et de robustesse :
1.	Garantie d’ordre “create → start”
•	runc create lance init qui prépare l’environnement puis se bloque sur open(O_WRONLY) du FIFO.
•	Tant que l’utilisateur n’a pas fait runc start (le reader), l’init ne peut pas avancer.
→ Ça force exactement la sémantique OCI : le conteneur ne démarre pas tant que start n’a pas été appelé.
2.	Simplicité côté host (Start)
•	Être reader est trivial : open(O_RDONLY) puis read(1).
•	Une fois l’octet lu, start peut unlink et sortir, sans avoir à garder un writer “keeper” ouvert pour éviter EOF, ni gérer SIGPIPE/EPIPE.
3.	Pas d’écriture host → container
•	Le host n’envoie aucune donnée dans le conteneur ; c’est l’init qui “pousse” un octet vers l’extérieur.
•	C’est plus sûr (évite toute ambiguïté d’injection) et correspond à l’idée “l’init notifie qu’il est prêt”.
4.	Comportement de blocage souhaité
•	open d’un FIFO en écriture seule bloque sans lecteur → parfait pour retenir init jusqu’à start.
•	L’inverse marcherait aussi (reader bloqué sans writer), mais tu te retrouves côté host à devoir écrire et gérer la durée de vie du writer proprement.
5.	Nettoyage prévisible
•	Ici, un seul writer (init). Après avoir écrit 1 octet, init ferme, le reader (start) lit et peut unlink immédiatement : pas de writers fantômes.

Schéma rapide :

create:        [init] ... setup ... open FIFO (WRONLY) ──(bloque)──┐
start:                                                   open (RDONLY) → read "0" → unlink
init:            write "0" → close → continue exécution

Bref : l’inversion est possible, mais ce design minimise la complexité côté host, garantit l’ordre, et évite d’écrire dans le conteneur. C’est pour ça que runc fait Init=WRONLY et Start=RDONLY.