https://github.com/opencontainers/runc/security/advisories/GHSA-xr7r-f8xq-vfvv?utm_source=chatgpt.com

Oui 👍 ce que tu observes est normal, et c’est exactement ce qui se passe dans un conteneur tant que tu n’as pas fait de nettoyage des descripteurs :

⸻

🔹 Ce que tu vois dans /proc/self/fd
•	0, 1, 2 → ton stdin, stdout, stderr reliés à /dev/pts/0 (le terminal du conteneur).
•	3, 4 → des pipe:[…] créés par le runtime pour la comm entre parent/enfant (ex : sync pipes, log pipe…).
•	5 → ton fameux exec.fifo qui a été passé exprès à l’init pour la synchro create/start.

C’est donc attendu : tu es en train de déboguer le process init du conteneur ou un process lancé dedans, donc tu vois les FDs que runc (ou ton code) lui a explicitement laissés ouverts.

⸻

🔹 Isolation et FDs
•	pivot_root / mount namespaces protègent les chemins, mais ne ferment pas les descripteurs déjà ouverts.
•	Donc si tu laisses des pipes, fifos, sockets ouverts au moment du execve, ils sont visibles dans /proc/self/fd et utilisables dans le conteneur.
•	C’est normal pour la synchro (ex : exec.fifo) → mais si tu veux de l’isolement “hermétique”, il faut fermer tout le reste.

⸻

🔹 Ce que fait runc
•	runc passe volontairement certains FDs (exec.fifo, log pipe, console socket…) à l’init.
•	Tout le reste est fermé avant le execve final (ils font un sweep sur les FDs hérités).
•	Résultat : un process utilisateur dans le conteneur ne voit que les FDs “autorisés”.

⸻

🔹 Dans ton cas
•	Tu vois pipe:[…] et /tmp/state/exec.fifo parce que ton init est lancé avant cleanup, ou bien parce que ton code ne ferme pas ces FDs.
•	C’est utile pour debugger, mais en “prod” tu voudrais :
•	limiter ce que tu gardes ouvert (typiquement 0/1/2 + les pipes de synchro dont tu as besoin),
•	fermer le reste juste avant execve.

⸻

✅ Donc oui, c’est “normal” de voir ces FDs dans /proc/self/fd du conteneur.
Ce n’est pas une faille, c’est un choix : tu leur donnes volontairement accès (par héritage).

⸻

Veux-tu que je t’indique le bout de code dans runc où ils font le “close-on-exec sweep” pour fermer tous les FDs inutiles avant de lancer le process final ?

----

Yes — runc ferme explicitement (et très tard) les FDs “internes” juste avant le execve() de l’init afin qu’aucun descripteur inutile ne fuite dans le process utilisateur.

Où ça se passe dans le code
•	La logique est centralisée dans libcontainer/utils avec la fonction UnsafeCloseFrom(minFd int) qui “balaye” et ferme tous les FDs ≥ minFd, en évitant ceux critiques au runtime Go. C’est exactement l’utilitaire utilisé par runc pour faire le grand ménage avant execve().  ￼
•	Suite aux vulnérabilités liées aux FDs hérités (ex. CVE-2024-21626), runc a durci le flux : fermeture de tous les FDs internes au dernier moment dans runc init, juste avant l’execve, pour empêcher leur réutilisation. Les notes de sécu et de release en parlent explicitement.  ￼

Comment runc s’en sert (résumé du pattern)
1.	Pendant create, runc ouvre certains FDs nécessaires (ex : exec.fifo en O_PATH|O_CLOEXEC) et passe exactement ceux voulus à l’init via ExtraFiles et des variables d’env (ex : _LIBCONTAINER_FIFOFD).
2.	Dans runc init, juste avant de lancer le process utilisateur, runc :
•	rouvre les FDs voulus via /proc/self/fd/N avec les bons flags I/O (pour le handshake),
•	puis appelle la routine de close-from pour fermer tout le reste (FD sweep),
•	enfin appelle execve().

Ainsi, seuls les FDs intentionnellement conservés (stdin/stdout/stderr + ceux strictement nécessaires comme le fifo de synchro) survivent à l’execve().

Si tu veux calquer ça dans ton implémentation :
•	garde ta liste blanche de FDs utiles,
•	appelle un équivalent de UnsafeCloseFrom(3) (ou close_range(3, ~0) si tu fais du C) juste avant exec,
•	et passe les FDs voulus via Cmd.ExtraFiles (Go enlèvera CLOEXEC uniquement pour eux).  ￼

Besoin que je t’indique un snippet Go “avant exec” qui ferme tous les FDs ≥ 3 sauf une whitelist (avec unix.Close en boucle sur /proc/self/fd) ?


----


Probable que FD 3 soit le premier FD passé via ExtraFiles (en Go, les FDs d’ExtraFiles commencent à 3, puis 4, etc.). Si tu as récemment passé exec.fifo ou un pipe au process, 3 = celui-là.

Voici comment l’identifier à coup sûr :

Rapide en shell (dans le conteneur)
1.	Voir les méta-infos du FD (type, flags)

cat /proc/self/fdinfo/3

	•	flags: est en octal sur Linux. Repères utiles :
	•	000000 = O_RDONLY, 000001 = O_WRONLY, 000002 = O_RDWR
	•	010000000 = O_PATH (si présent, c’est un handle sans I/O)
	•	000400000 = O_NONBLOCK
	•	pos: (offset) sera souvent 0 pour FIFO/pipe/socket.

	2.	Forcer un stat de la cible (même si le symlink n’est pas lisible)

stat -Lc 'mode=%f type=%F inode=%i' /proc/self/fd/3

	•	type te dira : fifo, pipe, socket, regular file, char device, etc.
	•	Si c’est un fifo/pipe tu verras fifo ou pipe.

	3.	Lister ce que le process a ouvert (si lsof dispo)

lsof -p $$ -nP | awk '$4 ~ /^3[urw]$/'

Mini utilitaire Go (identifie clairement le type)

package main

import (
"fmt"
"os"
"golang.org/x/sys/unix"
)

func main() {
fd := 3
var st unix.Stat_t
if err := unix.Fstat(fd, &st); err != nil {
fmt.Println("Fstat:", err)
return
}

	mode := st.Mode & unix.S_IFMT
	switch mode {
	case unix.S_IFREG:
		fmt.Println("FD 3: regular file")
	case unix.S_IFDIR:
		fmt.Println("FD 3: directory")
	case unix.S_IFLNK:
		fmt.Println("FD 3: symlink")
	case unix.S_IFCHR:
		fmt.Println("FD 3: char device")
	case unix.S_IFBLK:
		fmt.Println("FD 3: block device")
	case unix.S_IFIFO:
		fmt.Println("FD 3: FIFO (named pipe)")
	case unix.S_IFSOCK:
		fmt.Println("FD 3: socket")
	default:
		fmt.Printf("FD 3: unknown (mode=0%o)\n", st.Mode)
	}

	// Essaye de lire la cible du lien pour info (peut échouer avec O_PATH / unlink)
	target, err := os.Readlink("/proc/self/fd/3")
	fmt.Println("readlink:", target, err)
}

Interprétation fréquente
•	Si fdinfo montre flags: 010000000 → FD ouvert en O_PATH (typique quand tu as reproduit le pattern runc).
→ On ne peut pas lire/écrire dessus. Pour faire de l’I/O, réouvre via :

rd, err := os.OpenFile("/proc/self/fd/3", os.O_RDONLY, 0) // lecteur
// ou:
// fd, _ := unix.Open("/proc/self/fd/3", unix.O_WRONLY|unix.O_CLOEXEC, 0) // écrivain

Ça crée un nouveau FD utilisable (4, 5, …).

	•	Si type=fifo → c’est sûrement ton exec.fifo.
	•	Dans le protocole runc-like : init = writer, start = reader.
	•	Si type=pipe (pipe:[NNN]) → un pipe anonyme interne (logs, sync, etc.).

Rappel
•	Ne te fie pas au nom du symlink : pour un FD ouvert avec O_PATH ou après unlink, readlink peut échouer (ce que tu as vu : “cannot read link”). Ça ne veut pas dire que le FD est cassé ; juste qu’il n’a pas de chemin affichable.

Si tu colles ici la sortie de cat /proc/self/fdinfo/3 et stat -Lc ... /proc/self/fd/3, je te dis exactement ce que c’est et comment l’utiliser (lecture/écriture) dans ton flow.