Très bonne question 👌 Tu touches à un détail historique et architectural de Linux : pourquoi les PTY “modernes” (les slaves) vivent dans un FS spécial /dev/pts/, alors que les autres devices sont des nœuds statiques dans /dev.

⸻

🔹 1. Les anciens PTY (avant devpts)

Historiquement, Unix gérait les PTY avec des paires statiques dans /dev :
•	/dev/pty[p-za-e][0-9a-f] (master)
•	/dev/tty[p-za-e][0-9a-f] (slave)

Exemple : /dev/ttyp0 ↔ /dev/ptyp0.
👉 Problème : nombre limité, statique, compliqué à gérer.

⸻

🔹 2. Les PTY modernes : devpts

Linux a introduit un pseudo-fs spécial devpts :
•	Monté sur /dev/pts (par convention).
•	Chaque fois qu’un PTY est créé via posix_openpt() / openpty(), le noyau crée dynamiquement un device node dans ce FS.
•	Exemple : /dev/pts/0, /dev/pts/1, …

👉 Avantage : plus besoin de pré-créer des centaines de nœuds PTY statiques dans /dev.

⸻

🔹 3. Pourquoi un FS séparé et un sous-dossier ?
•	Isolation : le nombre de PTY peut exploser (un par conteneur, ssh session, tmux pane…). Les mettre dans /dev/pts évite d’encombrer /dev avec des centaines de fichiers.
•	Dynamisme : contrairement à /dev/null, /dev/zero etc. qui sont fixes, les PTY sont créés/supprimés dynamiquement → ça colle mieux avec un filesystem virtuel que des nodes statiques.
•	Permissions/namespaces :
•	En conteneur, chaque instance peut monter son propre devpts → isolation des PTY.
•	Ça permet d’avoir /dev/pts/0 dans deux conteneurs différents, sans collision.
•	Compatibilité : on a gardé /dev/tty (alias vers le ctty courant) pour ne pas casser le monde Unix classique, mais les PTY esclaves dynamiques vivent dans /dev/pts/.

⸻

🔹 4. En résumé
•	Les devices fixes (/dev/null, /dev/random, disques…) sont créés une fois pour toutes → nodes statiques dans /dev.
•	Les PTY esclaves sont éphémères et nombreux → gérés par un pseudo-fs spécial devpts monté sur /dev/pts.

⸻

🔹 Schéma ASCII

/dev (tmpfs ou bind)
│
├── null, zero, random, urandom, tty, console   ← devices fixes (via mknod)
│
└── pts/  (montage de devpts)
├── 0   ← PTY slave créé dynamiquement
├── 1
└── ...


⸻

👉 Donc : les PTY esclaves sont toujours sous /dev/pts/N, car devpts est un pseudo-fs fait exprès pour les gérer dynamiquement et isoler leur espace de noms.

Veux-tu que je t’explique aussi comment /dev/ptmx (le multiplexeur maître) est lié à devpts et sert à créer ces /dev/pts/N ?