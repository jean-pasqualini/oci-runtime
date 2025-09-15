Excellente question 👌 tu touches à un point central de la conception d’un conteneur : faut-il monter /dev à l’intérieur ?

⸻

🔹 Ce que contient “/dev” normalement

Dans un vrai système Linux, /dev contient :
•	les TTYs : /dev/tty, /dev/pts/*, /dev/console…
•	les devices système : /dev/null, /dev/zero, /dev/random, /dev/urandom, /dev/full…
•	les périphériques matériels (disques, block devices, etc.).

⸻

🔹 Dans un conteneur
•	La plupart des applications ont besoin d’un minimum de /dev pour tourner correctement :
•	/dev/null, /dev/zero, /dev/urandom (randomness, redirections, etc.).
•	/dev/tty si c’est un conteneur interactif.
•	/dev/pts si tu veux du vrai TTY multi-process (job control, etc.).
•	Mais tu ne veux pas exposer tout /dev de l’hôte (disques, GPU, etc.) → sinon le conteneur a accès à tes périphériques bruts 😱.

⸻

🔹 Comment font les runtimes (runc, docker, containerd)
•	Ils créent un tmpfs monté sur /dev dans le conteneur.
•	Puis ils ajoutent dedans uniquement les devices sûrs/minimaux via des device nodes et/ou mknod:
•	/dev/null
•	/dev/zero
•	/dev/full
•	/dev/random
•	/dev/urandom
•	/dev/tty (selon config)
•	Et éventuellement un devpts monté sur /dev/pts pour gérer les TTY interactifs.

👉 C’est exactement ce que fait Docker par défaut.

⸻

🔹 En résumé
•	Oui, tu es censé fournir un /dev dans ton conteneur.
•	Non, tu ne dois pas bind-mount /dev de l’hôte (sauf cas très spéciaux comme accès au GPU avec --device).
•	La bonne pratique → monter un tmpfs /dev avec les pseudo-devices nécessaires et éventuellement devpts si tu veux un terminal interactif.

⸻

✅ Donc :
•	Conteneur batch → /dev/null, /dev/zero, /dev/urandom suffisent.
•	Conteneur interactif → ajoute /dev/tty et /dev/pts.
•	Jamais tout le /dev de l’hôte.

⸻

👉 Tu veux que je te montre la liste exacte des mknod que fait runc pour initialiser un /dev minimal (celle qu’on retrouve dans la spec OCI) ?

----

Bonne question 👍 là on rentre dans la “plomberie Unix/Linux” de bas niveau.

⸻

🔹 Qu’est-ce qu’un device node ?
•	Ce sont des fichiers spéciaux dans /dev qui servent de porte d’entrée vers un pilote du noyau.
•	Exemple :
•	/dev/null → jette tout ce qu’on écrit dedans, lit toujours EOF.
•	/dev/zero → lit une infinité de \0.
•	/dev/random et /dev/urandom → fournissent de l’entropie (aléatoire) depuis le noyau.
•	Ce ne sont pas de vrais fichiers sur disque : ce sont des entrées du device filesystem qui pointent vers un major/minor number géré par le kernel.

⸻

🔹 Et mknod dans tout ça ?

mknod est l’appel système (et la commande) qui sert à créer un device node.
•	En C :

int mknod(const char *pathname, mode_t mode, dev_t dev);


	•	En shell :

mknod /dev/null c 1 3
mknod /dev/zero c 1 5
mknod /dev/random c 1 8
mknod /dev/urandom c 1 9



Ici :
•	c = character device (par opposition à b = block device).
•	Les deux chiffres (major, minor) identifient le driver dans le noyau.
•	Exemple /dev/null → major 1, minor 3.

⸻

🔹 Exemple pratique

rm -f /dev/null
mknod /dev/null c 1 3
chmod 666 /dev/null

→ tu recrées /dev/null.

⸻

🔹 Dans les conteneurs
•	Quand un runtime (comme runc) prépare /dev dans un conteneur, il monte un tmpfs vide et fait des mknod pour recréer les devices de base (null, zero, random, urandom, tty).
•	C’est pour ça qu’ils apparaissent dans ton conteneur même si /dev de l’hôte n’est pas monté.

⸻

🔹 Résumé
•	Device node = entrée dans /dev qui pointe vers un pilote du noyau.
•	mknod = appel système/commande pour créer ces entrées (avec type c/b et couple major/minor).
•	Exemples :
•	/dev/null = mknod /dev/null c 1 3
•	/dev/zero = mknod /dev/zero c 1 5
•	/dev/random = mknod /dev/random c 1 8
•	/dev/urandom = mknod /dev/urandom c 1 9

⸻

👉 Veux-tu que je te fasse un tableau avec tous les device nodes minimaux que runc/Docker recréent dans un conteneur standard (les noms + leurs major:minor + leur utilité) ?

-----

Parfait 🚀 Voici le tableau des device nodes minimaux qu’un conteneur (via runc/Docker) recrée dans son /dev.

⸻

🔹 Device nodes de base dans un conteneur

Device	Type	Major:Minor	Commande mknod équivalente	Utilité
/dev/null	c	1:3	mknod /dev/null c 1 3	Tout ce qu’on écrit est jeté ; lire renvoie EOF immédiat.
/dev/zero	c	1:5	mknod /dev/zero c 1 5	Fournit des octets \0 à l’infini (utile pour init de buffers).
/dev/full	c	1:7	mknod /dev/full c 1 7	Comme /dev/null, mais écrire échoue toujours avec ENOSPC.
/dev/random	c	1:8	mknod /dev/random c 1 8	Générateur de nombres aléatoires bloquant (attend l’entropie).
/dev/urandom	c	1:9	mknod /dev/urandom c 1 9	Générateur de nombres aléatoires non-bloquant.
/dev/tty	c	5:0	mknod /dev/tty c 5 0	Alias dynamique vers le controlling terminal du process.


⸻

🔹 Et côté TTY/PTYS
•	/dev/pts/ : monté via un devpts fs → chaque process interactif dans un conteneur peut avoir son PTY esclave (/dev/pts/N).
•	/dev/console : parfois recréé si la spec OCI l’exige, major:minor 5:1.

⸻

🔹 Exemple pratique en runtime

Ce que fait typiquement runc pour /dev :

mount -t tmpfs tmpfs /dev
mknod /dev/null c 1 3
mknod /dev/zero c 1 5
mknod /dev/full c 1 7
mknod /dev/random c 1 8
mknod /dev/urandom c 1 9
mknod /dev/tty c 5 0
mount -t devpts devpts /dev/pts


⸻

✅ Avec ça :
•	Les applis ont tous les devices “de base” pour tourner.
•	Tu n’exposes pas /dev/sda, /dev/nvme0n1, /dev/kvm, etc. de l’hôte → sécurité préservée.

⸻

👉 Veux-tu que je t’explique aussi pourquoi /dev/full est utile (même si on l’utilise rarement) et pourquoi il est inclus dans le set minimal ?