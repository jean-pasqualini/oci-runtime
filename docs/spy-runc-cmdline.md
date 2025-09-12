[2025-09-09 19:15:45] /usr/bin/runc --root /var/run/docker/runtime-runc/moby --log /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/2368c4208c99891477099303baed1a75d6f9aaffc6ae9d9b02deb2a450eb2300/log.json --log-format json exec --process /tmp/runc-process640585853 --detach --pid-file /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/2368c4208c99891477099303baed1a75d6f9aaffc6ae9d9b02deb2a450eb2300/ca2935fb9c611bd09ef839c550d53a73ed7224962991f7dc57a15a284a3fc84c.pid 2368c4208c99891477099303baed1a75d6f9aaffc6ae9d9b02deb2a450eb2300
[2025-09-09 19:15:45] /usr/bin/runc --root /var/run/docker/runtime-runc/moby --log /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/97e4fe2ead70ba0068244e25eaa56c0a7e2440d6335a775dcfa300c9a0a0b9d0/log.json --log-format json exec --process /tmp/runc-process3450212467 --detach --pid-file /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/97e4fe2ead70ba0068244e25eaa56c0a7e2440d6335a775dcfa300c9a0a0b9d0/5d02ebd1a41c94f32b25c8a0f83e202a61621926c66c35a3097b6a3eae4eda2a.pid 97e4fe2ead70ba0068244e25eaa56c0a7e2440d6335a775dcfa300c9a0a0b9d0
[2025-09-09 19:15:51] /usr/bin/runc --version
[2025-09-09 19:15:52] /usr/bin/runc --root /var/run/docker/runtime-runc/moby --log /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/fdc3fb1c0883f336fd3164ab130f2fd32c888e09f9e8af73cec027571f1b0c32/log.json --log-format json create --bundle /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/fdc3fb1c0883f336fd3164ab130f2fd32c888e09f9e8af73cec027571f1b0c32 --pid-file /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/fdc3fb1c0883f336fd3164ab130f2fd32c888e09f9e8af73cec027571f1b0c32/init.pid --console-socket /tmp/pty2578984118/pty.sock fdc3fb1c0883f336fd3164ab130f2fd32c888e09f9e8af73cec027571f1b0c32
[2025-09-09 19:15:52] /usr/bin/runc --root /var/run/docker/runtime-runc/moby --log /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/fdc3fb1c0883f336fd3164ab130f2fd32c888e09f9e8af73cec027571f1b0c32/log.json --log-format json start fdc3fb1c0883f336fd3164ab130f2fd32c888e09f9e8af73cec027571f1b0c32

----

[2025-09-09 19:17:32] /usr/bin/runc --root /var/run/docker/runtime-runc/moby --log /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/af6198ff5bcffa5c6484adad6d3acc2beff9f298bf3a9d290b1b317b633fdfe4/log.json --log-format json create --bundle /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/af6198ff5bcffa5c6484adad6d3acc2beff9f298bf3a9d290b1b317b633fdfe4 --pid-file /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/af6198ff5bcffa5c6484adad6d3acc2beff9f298bf3a9d290b1b317b633fdfe4/init.pid --console-socket /tmp/pty199751582/pty.sock af6198ff5bcffa5c6484adad6d3acc2beff9f298bf3a9d290b1b317b633fdfe4
[2025-09-09 19:17:32] /usr/bin/runc --root /var/run/docker/runtime-runc/moby --log /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/af6198ff5bcffa5c6484adad6d3acc2beff9f298bf3a9d290b1b317b633fdfe4/log.json --log-format json start af6198ff5bcffa5c6484adad6d3acc2beff9f298bf3a9d290b1b317b633fdfe4


### Bash wrapper
```bash
#!/bin/bash
# chemin vers le vrai runc (à adapter si nécessaire)
REAL_RUNC="/usr/bin/runc-real"
LOGFILE="/tmp/runc-log.txt"

# logge la date + la ligne de commande complète
{
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $0 $@"
} >> "$LOGFILE"

# exécute le vrai runc avec tous les arguments
exec "$REAL_RUNC" "$@"
```

### Explain

Parfait 👍 tu as intercepté exactement ce que containerd-shim lance.
Décortiquons chaque morceau :

⸻

📌 Première commande (phase create)

/usr/bin/runc \
  --root /var/run/docker/runtime-runc/moby \
  --log /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/<id>/log.json \
  --log-format json \
  create \
  --bundle /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/<id> \
  --pid-file /var/run/desktop-containerd/daemon/io.containerd.runtime.v2.task/moby/<id>/init.pid \
  --console-socket /tmp/pty199751582/pty.sock \
  <id>

Que signifient ces options ?
	•	--root /var/run/docker/runtime-runc/moby
👉 Répertoire d’état où runc garde ses métadonnées (sockets, états de conteneur, etc.) pour ce namespace Docker Desktop.
	•	--log .../log.json + --log-format json
👉 Où écrire les logs d’exécution de runc (très utiles pour containerd-shim en cas d’erreur).
	•	create
👉 Dit à runc de créer le conteneur (mais ne pas encore démarrer le process init).
Il lit le bundle OCI (config.json + rootfs) depuis --bundle.
	•	--bundle .../<id>
👉 Chemin vers le répertoire du bundle OCI généré par containerd-shim.
Il contient :
	•	config.json (spec OCI du conteneur)
	•	rootfs/ (système de fichiers monté)
	•	init.pid (sera écrit plus bas)
	•	--pid-file .../init.pid
👉 Fichier où runc écrit le PID du process init du conteneur.
	•	--console-socket /tmp/.../pty.sock
👉 Socket UNIX temporaire par lequel runc remet un PTY maître au shim (pour un conteneur avec terminal: true).
Sans ça, ce serait des FIFOs (stdin, stdout, stderr).
	•	<id>
👉 L’ID unique du conteneur côté containerd (un hash qui correspond aussi à ce que tu vois dans docker ps).

⸻

📌 Deuxième commande (phase start)

/usr/bin/runc \
  --root /var/run/docker/runtime-runc/moby \
  --log .../log.json \
  --log-format json \
  start <id>

	•	Ici, runc reprend le conteneur créé à l’étape précédente (dans --root + <id>).
	•	start demande à runc de lancer le process init (défini dans config.json).
	•	Après ça, runc se termine : c’est le shim qui reste en vie, surveille le conteneur, et gère l’I/O + lifecycle.

⸻

🔄 Résumé du cycle
	1.	runc create ...
→ installe le conteneur dans les namespaces, cgroups, rootfs, mais process init pas encore exécuté.
	2.	runc start <id>
→ lance le process init.
Après ça, runc sort.
Le shim reste pour garder la session ouverte et fournir les logs/I/O.

⸻

👉 En clair, ton log montre le pattern classique containerd v2 :
	•	create avec beaucoup d’options techniques (--bundle, --pid-file, --console-socket)
	•	puis start avec juste --root et l’id.

https://github.com/opencontainers/runc/blob/main/libcontainer/container_linux.go
⸻

Veux-tu que je te montre comment retrouver le config.json exact (le spec OCI) de ce conteneur dans --bundle, pour voir tout ce que runc reçoit comme configuration ?
https://chromium.googlesource.com/external/github.com/docker/containerd/%2B/refs/tags/v0.2.1/vendor/src/github.com/opencontainers/runc/libcontainer/container_linux.go?utm_source=chatgpt.com
```bash
lsns | wc -l
wget https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/aarch64/alpine-minirootfs-3.19.1-aarch64.tar.gz
mkdir -p /tmp/mycontainer/rootfs
tar -C /tmp/mycontainer/rootfs -xzf alpine-minirootfs-3.19.1-aarch64.tar.gz
sed -i 's/"terminal": true/"terminal": false/' config.json
```

--------


Oui, exec.fifo passe aussi par cmd.ExtraFiles ✅

Concrètement, côté runc create :
•	runc crée la FIFO nommée (…/exec.fifo) sur disque.
•	Il l’ouvre et place le FD correspondant dans cmd.ExtraFiles avant d’exec le sous-processus runc init.
•	Il expose le numéro de ce FD à l’init via l’ENV _LIBCONTAINER_FIFOFD=<n>.

Côté runc init :
•	L’init lit _LIBCONTAINER_FIFOFD, récupère ce FD (visible aussi comme /proc/self/fd/<n>), et s’y bloque en lecture jusqu’à ce que runc start ouvre l’autre extrémité (écriture). C’est exactement ce qu’on voit dans les traces où runc init est bloqué sur /proc/self/fd/5 → lien symbolique vers exec.fifo.  ￼ ￼

Donc, résumé :
•	Config JSON → pipe anonyme passé via _LIBCONTAINER_INITPIPE (FD dans ExtraFiles, lu une fois puis fermé).
•	Synchro de démarrage (exec.fifo) → FD aussi passé via ExtraFiles, référencé par _LIBCONTAINER_FIFOFD, sur lequel runc init se bloque jusqu’au start.  ￼ ￼

Si tu veux, je peux te pointer les lignes précises dans le code (sendConfig() / standard_init_linux.go) où ces ENV et FDs sont câblés.

---------

Exactement 👍

Dans runc il y a deux mécanismes distincts, donc deux variables d’environnement :

⸻

🔹 _LIBCONTAINER_INITPIPE
•	Quoi ? un pipe anonyme créé par le parent (runc create).
•	Passage via : cmd.ExtraFiles → FD transmis à l’enfant.
•	But : transporter le JSON de configuration (configs.Config) du parent vers runc init.
•	Cycle de vie : utilisé une seule fois (read + unmarshal), puis fermé.

⸻

🔹 _LIBCONTAINER_FIFOFD
•	Quoi ? la FIFO nommée exec.fifo créée sur disque.
•	Passage via : son FD aussi placé dans cmd.ExtraFiles, et exposé par cette variable.
•	But : blocage/synchronisation → runc init lit dessus, et attend que runc start ouvre l’autre extrémité en écriture.
•	Cycle de vie : reste ouvert tant que le conteneur n’est pas “starté”.

⸻

👉 Donc oui : INITPIPE = passage de config JSON, FIFOFD = synchro pour le start.

Tu veux que je t’écrive un petit tableau chronologique (qui ouvre quoi, qui ferme quoi, à quel moment) pour bien voir comment ces deux FDs vivent au cours de create → init → start ?


-----------

Parfait — voilà le timeline FD → action clair et net pour runc seul (et c’est identique côté FDs quand containerd-shim est là, à part le parent adoptif).

Chronologie des FDs (INITPIPE & FIFOFD)

Phase	Qui	Action sur FD	Détails
0) Prépare (runc create, parent)	runc (parent)	Crée un pipe anonyme → extrémité enfant passée via cmd.ExtraFiles	Expose l’index via _LIBCONTAINER_INITPIPE=<n>. Servira à envoyer le JSON de config à runc init.
   Crée la FIFO nommée …/exec.fifo et l’ouvre	Met ce FD aussi dans cmd.ExtraFiles et expose _LIBCONTAINER_FIFOFD=<m>.
1) Fork/exec (runc init)	runc (parent)	Exec du même binaire avec l’arg init	Les FDs d’ExtraFiles deviennent accessibles dans l’enfant (en général dès FD 3+).
2) Bootstrap config	runc init (enfant)	Ouvre/dup le FD _LIBCONTAINER_INITPIPE en lecture	Lit TOUT le JSON (config libcontainer : Rootfs, mounts, namespaces, cgroups, process.args/env…).
   runc (parent)	Écrit le JSON dans _LIBCONTAINER_INITPIPE puis ferme	Usage one-shot. Si l’enfant meurt → parent voit “broken pipe”.
3) Setup conteneur	runc init	Applique la config (mounts, pivot_root, setns, cgroups, uid/gid, caps…)	À ce stade, config déjà reçue → INITPIPE n’est plus utilisé.
4) Synchronisation Start	runc init	Ouvre _LIBCONTAINER_FIFOFD en lecture et se bloque	Le lien /proc/self/fd/<m> pointe vers …/exec.fifo. Tant que pas d’open en écriture, bloqué.
   runc create (parent)	Sort	Sans shim : l’init sera adopté par PID 1 après le start. Avec shim : adopté par le shim (subreaper).
5) Start	runc start (nouveau proc)	Ouvre exec.fifo en écriture	Ce simple open suffit à débloquer la lecture côté init (même sans écrire).
6) Lancement entrypoint	runc init	Débloqué, puis execve(process.args)	Ton vrai process remplace runc init. La FIFO peut être fermée/retirée.
7) Vie du conteneur	process user (entrypoint)	—	IO/Signals gérés par ton environnement (PID 1 ou shim). runc n’est plus là.


⸻

Points clés à retenir
•	Deux canaux distincts :
•	_LIBCONTAINER_INITPIPE → pipe anonyme pour transmettre la config JSON (one-shot, fermé après lecture/écriture).
•	_LIBCONTAINER_FIFOFD → FIFO nommée pour la synchro de démarrage (init lit/bloque ; runc start ouvre en écriture pour le réveiller).
•	Aucun passage de config via la FIFO (elle sert uniquement à la synchro “created → started”).
•	Sans shim : après start, le process conteneur est ré-adopté par PID 1 de l’host. Avec shim : il est ré-adopté par le shim (subreaper).

⸻

Comment l’observer concrètement
•	Voir la FIFO et qui la tient :

# Après `runc create <id>`
ls -l /run/runc/<id>/exec.fifo        # (rootless: $XDG_RUNTIME_DIR/runc/<id>/)
sudo lsof /run/runc/<id>/exec.fifo    # montre le lecteur (runc:[2:INIT])


	•	Lier un FD à la FIFO :

pid=$(ps -ef | grep "runc:\[2:INIT\]" | grep -v grep | awk '{print $2}')
sudo ls -l /proc/$pid/fd              # l’un des FD pointe vers .../exec.fifo


	•	Tracer le déblocage au start :

sudo strace -p $pid -e read,openat -f
# Dans un autre terminal: runc start <id>
# → tu verras la lecture se débloquer puis l'execve de l’entrypoint.



Tu veux la version “avec shim” sur le même schéma (avec adoption, logpipe, stdio, signaux) ? Je te le fais dans la foulée.
-------

pipes

Parfait 👌 je vais te donner la liste complète des variables d’environnement _LIBCONTAINER_* que runc utilise pour “documenter” les FDs qu’il passe à runc init.

⚠️ Attention : certaines sont toujours utilisées, d’autres sont optionnelles (selon les features activées : console, systemd, pidfd, etc.).

⸻

📜 Liste des _LIBCONTAINER_*

🔹 Communication & synchro
•	_LIBCONTAINER_INITPIPE
•	Pipe anonyme parent → enfant.
•	Sert à transférer le JSON de configuration (configs.Config).
•	One-shot, fermé après lecture.
•	_LIBCONTAINER_SYNCPIPE
•	Pipe/sockpair pour handshake bootstrap.
•	L’enfant (runc init) signale au parent (runc create) l’état d’avancement (pivot_root, erreurs).
•	Fermé une fois l’init prêt.
•	_LIBCONTAINER_FIFOFD
•	FD vers la FIFO nommée exec.fifo.
•	L’init lit dessus et se bloque.
•	Débloqué par runc start qui ouvre l’autre côté.
•	_LIBCONTAINER_LOGPIPE
•	Pipe de logs.
•	L’enfant écrit ses erreurs bootstrap dedans.
•	Le parent (runc) les lit pour remonter proprement à l’utilisateur.
•	Optionnellement contrôlé par _LIBCONTAINER_LOGLEVEL.

⸻

🔹 Console / TTY
•	_LIBCONTAINER_CONSOLE
•	FD vers un socket Unix permettant de passer le master d’un pty.
•	Sert quand tu veux une console TTY attachée au conteneur.
•	_LIBCONTAINER_CONSOLE_SOCKET (ancien / alias)
•	Dans certaines versions, même usage que ci-dessus.

⸻

🔹 Processus / PID
•	_LIBCONTAINER_PIDFILE
•	Chemin d’un fichier où écrire le PID de l’init (quand demandé avec --pid-file).
•	_LIBCONTAINER_PIDFD_SOCK
•	Si le noyau supporte pidfd, runc passe un socket pour récupérer un pidfd de l’init.
•	Permet un suivi sûr (évite la réutilisation de PID).

⸻

🔹 Cgroups / Systemd
•	_LIBCONTAINER_CGROUPS (rare, dépend du code interne)
•	Peut contenir un FD vers un socket utilisé pour gérer systemd-cgroup manager.
•	_LIBCONTAINER_SYSTEMD_CGROUP
•	Variable booléenne (true/false) : indique si cgroups sont gérés via systemd.

⸻

🔹 Divers
•	_LIBCONTAINER_LOGLEVEL
•	Niveau de logs pour l’enfant (error, warn, info, debug).
•	Contrôle ce que l’enfant écrit dans LOGPIPE.

⸻

✅ Résumé (les “incontournables”)

Dans presque tous les cas, tu verras au moins :
•	_LIBCONTAINER_INITPIPE
•	_LIBCONTAINER_SYNCPIPE
•	_LIBCONTAINER_FIFOFD
•	_LIBCONTAINER_LOGPIPE
(+ éventuellement _LIBCONTAINER_CONSOLE si TTY, _LIBCONTAINER_PIDFD_SOCK si pidfd).

⸻

👉 Veux-tu que je fasse un tableau (colonne = variable, FD ou path associé, rôle, durée de vie) pour avoir un cheat sheet complet ?