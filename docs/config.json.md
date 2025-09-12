Top—down, champ par champ. Je commente aussi les subtilités (unités, LSM, cgroups v1/v2, formats “anciens” vs “modernes”).

En-tête & plate-forme
•	ociVersion: "0.5.0-dev"
Version du schéma OCI attendue par le runtime. Ici très ancienne (pré-1.0). Certains champs (ex. process.capabilities en simple liste) reflètent cet ancien format.
•	platform.os: "linux" / platform.arch: "amd64"
Plateforme visée pour l’exécution (utile pour des bundles multi-plateformes).

Processus (le “PID 1” dans le conteneur)
•	process.terminal: true
Attache un pseudo-terminal (pty). Le runtime crée typiquement un devpts et connecte stdio au pty maître.
•	process.user
•	uid: 1, gid: 1 → UID/GID effectifs dans le conteneur (après éventuel remap userns).
•	additionalGids: [5,6] → groupes suppl. (ex. tty=5, disk=6 selon distro).
•	process.args: ["sh"]
Programme + arguments à exécuter (argv[0] = sh ici).
•	process.env: ["PATH=…", "TERM=xterm"]
Variables d’environnement initiales.
•	process.cwd: "/"
Répertoire de travail du process au démarrage.
•	process.capabilities: ["CAP_AUDIT_WRITE","CAP_KILL","CAP_NET_BIND_SERVICE"]
Ancien format: une seule liste. Les OCI modernes séparent en bounding/effective/permitted/inheritable/ambient. Ici on accorde ces 3 capacités de base.
•	process.rlimits
Ulimits POSIX à appliquer :
•	RLIMIT_CORE {hard:1024, soft:1024} → taille max des core dumps (en blocs sur certaines implémentations/ABI).
•	RLIMIT_NOFILE {hard:1024, soft:1024} → nombre max de descripteurs ouverts.
•	process.apparmorProfile: "acme_secure_profile"
Nom du profil AppArmor à charger (nécessite AppArmor activé côté hôte).
•	process.selinuxLabel: "system_u:system_r:svirt_lxc_net_t:s0:c124,c675"
Contexte SELinux (confinement de processus) si SELinux est actif.
•	process.noNewPrivileges: true
Active PR_SET_NO_NEW_PRIVS: le process et ses enfants ne peuvent plus élever leurs privilèges (désactive effets setuid/setcap).

Système de fichiers racine
•	root.path: "rootfs"
Chemin (relatif au bundle) de la racine à monter dans le conteneur.
•	root.readonly: true
Monte / en lecture seule (les tmpfs/montages en écriture restent possibles ailleurs).

Identité UTS
•	hostname: "slartibartfast"
Nom d’hôte (dans le namespace UTS du conteneur).

Montages (fs “de base”)

Chaque entrée correspond à un mount(2) effectué dans le namespace de montage du conteneur :
•	/proc (type: "proc")
Pseudo-FS du noyau.
•	/dev (type: "tmpfs", options nosuid, strictatime, mode=755, size=65536k)
Espace de périphériques en tmpfs. mode=0755, taille 64 MiB.
•	/dev/pts (type: "devpts", newinstance, ptmxmode=0666, mode=0620, gid=5)
Espace de pseudo-tty privé. GID 5 (groupe tty), permissions ptmx 0666.
•	/dev/shm (type: "tmpfs", mode=1777, size=65536k)
Mémoire partagée POSIX, sticky bit (1777), 64 MiB.
•	/dev/mqueue (type: "mqueue")
Files de messages POSIX.
•	/sys (type: "sysfs")
Vue sysfs du noyau (ici nosuid,noexec,nodev).
•	/sys/fs/cgroup (type: "cgroup", options ro,relatime,…)
Montage hiérarchies cgroups v1 en lecture seule. Sur cgroup v2 modernes, on monterait type: "cgroup2".

Hooks (scripts appelés par le runtime)
•	hooks.prestart
Exécutés après création des namespaces mais avant le démarrage du process (sémantique fine a évolué dans les versions ultérieures).
•	/usr/bin/fix-mounts arg1 arg2, avec env key1=value1
•	/usr/bin/setup-network
•	hooks.poststart
Après que le process a démarré.
•	/usr/bin/notify-start, timeout: 5 (secondes)
•	hooks.poststop
Après l’arrêt/clean du conteneur.
•	/usr/sbin/cleanup.sh -f

Spécifique Linux

Périphériques /dev explicites
•	linux.devices
Demande au runtime de créer ces nœuds dans /dev et (souvent) d’ouvrir l’accès côté cgroup devices :
•	/dev/fuse: type: "c" (caractère), major 10 minor 229, fileMode: 438 (= 0666), owned by root.
•	/dev/sda: type: "b" (bloc), major 8 minor 0, fileMode: 432 (= 0660).

User namespaces (mappages UID/GID)
•	uidMappings / gidMappings
•	hostID: 1000, containerID: 0, size: 32000
Mappe l’UID/GID 0..31999 dans le conteneur vers 1000..32999 sur l’hôte. Permet d’avoir “root dans le conteneur” sans être root sur l’hôte (rootless-friendly).

Sysctl (namespaced)
•	sysctl: {"net.ipv4.ip_forward":"1","net.core.somaxconn":"256"}
Applique des sysctl (doivent être autorisés/“namespaced” sinon rejetés).

cgroups
•	cgroupsPath: "/myRuntime/myContainer"
Chemin du groupe de contrôle.
•	Si relatif, le runtime le résout sous son “scope” par défaut.
•	Ici absolu: placé tel quel (sur v1: multiples hierarchies; sur v2: unique).
•	resources (quotas/limites de ressources)
Réseau (cgroups net_cls / net_prio, v1)
•	network.classID: 1048577 (0x00100001) → marquage pour tc/iptables.
•	network.priorities: eth0:500, eth1:1000 → priorité d’interface (v1 net_prio).
PIDs
•	pids.limit: 32771 → nb max de processus/threads dans le cgroup.
Huge pages
•	hugepageLimits: pageSize:"2MB", limit:9223372036854772000 → quota bytes pour pages 2MiB.
OOM
•	oomScoreAdj: 100 → impact sur le score OOM du kernel (plus grand = tué plus tôt).
•	disableOOMKiller: false → ne désactive pas le tueur OOM pour ce cgroup.
Mémoire (cgroup v1)
•	limit: 536870912 → hard limit (512 MiB).
•	reservation: 536870912 → soft limit (hint).
•	swap: 536870912 → limite mémoire+swap (v1); sur v2, syntaxe différente.
•	kernel: 0, kernelTCP: 0 → quotas mémoire noyau (0 = non fixé).
•	swappiness: 0 → propension à swapper (0 = évite).
CPU (CFS/RT)
•	shares: 1024 → poids relatif en contention.
•	quota: 1000000, period: 500000 → CFS: autorise 1s de CPU / 0,5s → 2 vCPU effectifs (quota/period).
•	realtimeRuntime: 950000, realtimePeriod: 1000000 → budgets CPU temps réel (nécessite RT + permissions).
•	cpus: "2-3" → affinité CPU (cpuset) aux cœurs 2 et 3.
•	mems: "0-7" → nœuds NUMA autorisés (cpuset.mems).
Devices (filtre cgroup devices)
•	Règle par défaut {"allow": false, "access": "rwm"} → tout interdit (deny-all).
•	Exceptions autorisées :
•	Caractère major 10 minor 229 (/dev/fuse) accès rw.
•	Bloc major 8 minor 0 (/dev/sda) accès r.
Block I/O (v1 blkio)
•	blkioWeight: 10, blkioLeafWeight: 10 → pondération faible (10/1000 typ.).
•	blkioWeightDevice → poids par périphérique (8:0 et 8:16).
•	blkioThrottleReadBpsDevice [{8:0, rate:600}] → 600 B/s lecture max sur 8:0.
•	blkioThrottleWriteIOPSDevice [{8:16, rate:300}] → 300 IOPS écriture max.
(Sur cgroup v2, on utiliserait io.max, io.weight etc.)
•	rootfsPropagation: "slave"
Flag de propagation sur / (empêche les montages extérieurs d’“inonder” la vue interne; rslave typiquement).

Seccomp
•	seccomp.defaultAction: "SCMP_ACT_ALLOW"
Politique par défaut = autoriser les syscalls.
•	seccomp.architectures: ["SCMP_ARCH_X86"]
ABIs concernées (x86 32-bit ici; pour amd64 on verrait souvent "SCMP_ARCH_X86_64" aussi).
•	seccomp.syscalls: [{name:"getcwd", action:"SCMP_ACT_ERRNO"}]
Bloque getcwd(2) en renvoyant une erreur (errno).

Namespaces
•	namespaces: [{"type":"pid"}, "network","ipc","uts","mount","user","cgroup"]
Liste des NS à isoler.
•	user → nécessaire si on utilise uidMappings/gidMappings.
•	cgroup ns isole la vue des cgroups.

Chemins /proc protégés
•	maskedPaths: ["/proc/kcore", "/proc/latency_stats", "/proc/timer_stats", "/proc/sched_debug"]
Masqués (montés sur tmpfs vide ou bind vers /dev/null) pour éviter la fuite d’infos/usage risqué.
•	readonlyPaths: ["/proc/asound","/proc/bus","/proc/fs","/proc/irq","/proc/sys","/proc/sysrq-trigger"]
Bind-mount en lecture seule pour ces chemins de contrôle sensibles.

SELinux label pour montages
•	mountLabel: "system_u:object_r:svirt_sandbox_file_t:s0:c715,c811"
Label appliqué aux montages (différent du label de process).

Annotations
•	annotations: {"com.example.key1":"value1","com.example.key2":"value2"}
Métadonnées libres (clé/valeur) pour orchestrateurs/outillage. Non interprétées par le kernel.

⸻

Notes de compat & pièges fréquents
•	Capacités: le format moderne attend un objet {bounding, permitted, effective, inheritable, ambient}. Ici c’est une liste unique (vieux schéma).
•	cgroups v2: beaucoup de champs memory/*, blockIO/*, network/* sont spécifiques v1. En v2, les clés changent (ex. io.max, memory.max, pids.max).
•	RT CPU: realtimeRuntime/Period ne marchent que si l’hôte autorise RT pour le cgroup (souvent désactivé).
•	Seccomp arch: sur amd64, ajoute généralement SCMP_ARCH_X86_64 (et parfois SCMP_ARCH_X32) pour couvrir toutes les ABIs.
•	/sys/fs/cgroup type: aujourd’hui, on monte souvent cgroup2 (unifié).
•	Devices: penser à aligner linux.devices (création des nœuds) et resources.devices (policy cgroup) sinon le nœud existe mais l’accès est refusé.

Si tu veux, je te refais ce même fichier en schéma OCI 1.1+ “moderne” (capabilities en sets, cgroup v2, io.*) avec des valeurs équivalentes.