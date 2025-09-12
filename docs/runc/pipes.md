Très bonne question ! 🙂 Quand tu utilises runc, il y a effectivement une mécanique assez précise autour des pipes entre le parent (runc) et le process enfant (le container init process). Je vais te détailler ça :

⸻

1. Quels pipes sont utilisés ?

runc crée plusieurs anonymous pipes (pas nommés dans le filesystem, juste ouverts entre parent/enfant avant exec).
Les principaux sont :
•	exec.fifo / exec.fifo pipe : sert à synchroniser le moment où le process init dans le container commence réellement.
•	parent <-> init pipe : un canal de contrôle entre runc (parent) et le process init qu’il a forké.
•	log pipe : pour que le process init puisse renvoyer ses erreurs ou logs structurés en JSON au parent.
•	user-defined pipes : par exemple quand tu utilises --console-socket pour la gestion du TTY, un socket/pipe supplémentaire est passé.

⸻

2. À quoi servent-ils ?

Chaque pipe a une fonction claire :
•	init pipe (aussi appelé “sync pipe”)
Sert à synchroniser la création du container.
→ Le parent écrit des commandes (ex. “set up namespaces”, “pivot_root done”, “ready to exec…”) et l’enfant répond “ok” ou renvoie un code d’erreur.
C’est comme un mini-protocole entre runc et runc init.
•	log pipe
Si l’init échoue (mauvais mount, pivot_root impossible…), l’enfant écrit une structure JSON avec msg, level, error.
Ça permet au parent d’afficher un message clair plutôt qu’un simple exit 1.
•	exec fifo
Petit verrou/signal : le process init s’endort sur un read() tant que le parent n’a pas écrit dedans.
→ Ça permet de préparer le container, configurer les cgroups, puis seulement débloquer l’init pour exécuter ton vrai process.
•	console-socket (optionnel)
Si tu as demandé un TTY (--console-socket), le parent passe un FD de socket via SCM_RIGHTS.
Sert à connecter le pty de l’hôte au container.

⸻

3. Sont-ils anonymes ?

Oui 👍
•	Ce sont des anonymous pipes créés avec pipe2() (donc pas visibles dans /tmp ou ailleurs).
•	Ils existent uniquement via leurs file descriptors hérités entre parent/enfant.
•	Quand les deux extrémités sont fermées, ils disparaissent.

⸻

4. Quelles infos sont transmises ? Quand ?
   •	Pendant le setup du container :
   Le parent (runc run) envoie au runc init les ordres de config (via le pipe init/sync).
   L’init répond quand chaque étape est prête (namespaces montés, pivot_root fait, chdir, etc.).
   •	En cas d’erreur :
   L’init écrit dans le log pipe un message JSON décrivant l’erreur. Le parent peut alors afficher l’erreur et nettoyer.
   •	Avant d’exécuter le vrai process utilisateur :
   Le parent attend que tout soit configuré (cgroups, seccomp, etc.), puis il débloque le exec fifo → le process init lance enfin ton binaire dans le container.
   •	Optionnellement :
   Si console TTY demandé → échange des FDs via le console-socket.

⸻

👉 En résumé :
•	Oui, ce sont des pipes anonymes.
•	Ils servent surtout à synchroniser, signaler les erreurs, et gérer la console.
•	Les infos transmises sont des petits protocoles maison (structs binaires ou JSON), pas des gros flux de données.
•	Tout ça se passe entre le moment où tu lances runc run et le moment où ton process container démarre vraiment.

⸻

Veux-tu que je te fasse un schéma chronologique (parent <-> init avec pipes, qui écrit quoi à quel moment) pour bien visualiser le déroulement ?


----


Bonne question — j’ai vérifié dans le code de runc ce qui circule exactement sur le “pipe de sync”.

Ce qui passe dans le pipe de sync

Le pipe de sync échange de tout petits messages JSON de la forme {"type": "<valeur>"} où <valeur> est l’un des types suivants :
•	procReady → envoyé par l’init (enfant) pour dire “je suis prêt, fais le setup final”.
•	procRun → envoyé par le parent pour donner le feu vert “tu peux exec maintenant”.
•	procHooks → envoyé par l’init (enfant) pour demander au parent d’exécuter les hooks pre-start (dans certains cas de namespaces).
•	procResume → envoyé par le parent pour dire “hooks terminés, reprends”.
•	procError + un second JSON genericError → utilisé par l’un ou l’autre côté pour signaler une erreur détaillée (le lecteur lit d’abord {"type":"procError"}, puis un objet erreur).

Il n’y a pas d’autre payload avec ces messages-là : juste ce petit JSON avec le champ type (sauf en cas d’erreur, où un objet genericError suit).

Qui envoie quoi et quand (résumé chronologique)
1.	L’init (processus enfant runc:[2:INIT]) termine sa préparation (rootfs, console, etc.), puis écrit procReady et attend procRun.
2.	Le parent, en recevant procReady, applique les derniers réglages sur le vrai PID host : cgroups (Set), oom_score_adj, rlimits, et (si pas de NEWNS) exécute les pre-start hooks, puis renvoie procRun.
3.	Si les hooks doivent être exécutés plus tard (cas avec NEWNS), l’init envoie procHooks, le parent exécute les hooks et répond procResume.
4.	En cas d’erreur, l’émetteur envoie procError suivi d’un objet genericError ; le récepteur lève l’erreur.

Détails utiles
•	Le gros blob de configuration (la initConfig) est envoyé avant ce handshake, via le même pipe mais comme un JSON complet distinct (pas un “type de sync”). C’est fait par le parent avec utils.WriteJSON(p.parentPipe, p.config).
•	La console/pty n’est pas transmise sur le pipe de sync : le FD est envoyé via un socket Unix séparé avec SCM_RIGHTS (utils.SendFd(...)). Les mentions historiques de procConsole* dans les commentaires ne correspondent plus à des types actifs sur le sync pipe.

En bref : le pipe de sync ne transporte que des signaux d’état ultra-simples (procReady, procRun, procHooks, procResume, procError + erreur) pour coordonner parent et init — rien d’autre.
------

Super question — oui, il y a procRun et un exec.fifo… mais ils ne servent pas au même niveau.

TL;DR
•	procRun = petit message JSON envoyé dans le pipe de sync (anonyme), parent → enfant, pour dire “OK, tu peux exec maintenant”. C’est intra-invocation (le même runc parent qui a forké l’init).
•	exec.fifo = FIFO nommée sur disque (/run/runc/<id>/exec.fifo) utilisée comme barrière de démarrage entre deux commandes distinctes (runc create puis plus tard runc start). L’init bloque dessus jusqu’à ce que runc start l’ouvre/écrive un octet, ce qui le débloque.  ￼

⸻

Pourquoi avoir exec.fifo si on a déjà procRun ?

Parce qu’ils couvrent deux cas d’usage différents :
1.	Chemin “one-shot” : runc run
•	Le parent envoie la grosse initConfig sur le pipe de sync, l’enfant prépare, envoie procReady.
•	Le parent finit ses réglages (cgroups, hooks…), puis envoie procRun → l’enfant fait execve().
•	Ici, exec.fifo est ouvert/consommé immédiatement par le même processus runc (il ne sert que de garde-fou interne).  ￼
2.	Chemin “split phase” : runc create puis runc start
•	runc create lance l’init qui se bloque en ouvrant/écrivant exec.fifo (personne en face → blocage).
•	Plus tard, une autre invocation (runc start) ouvre et lit/écrit ce FIFO → débloque l’init, qui peut alors recevoir procRun sur le pipe de sync et exec.
•	L’intérêt : permettre à un orchestrateur (docker, containerd, podman…) de créer d’abord le conteneur, de faire d’autres préparations, puis de déclencher le démarrage au bon moment.  ￼

⸻

Détails concrets
•	exec.fifo est une FIFO nommée (visible dans /run/runc/<id>/exec.fifo), pas un pipe anonyme. L’init s’y bloque sur open/write tant qu’aucun lecteur (runc start) n’est là. Quand runc start l’ouvre et lit/écrit un octet (la valeur importe peu), l’init est débloqué et la suite du handshake (dont procRun) peut se conclure.  ￼
•	Ce mécanisme existe notamment pour éviter des courses entre start/delete et garantir qu’on ne démarre pas “trop tôt” ou qu’on ne bloque pas indéfiniment — il y a même eu des correctifs autour de cette FIFO.  ￼

⸻

En une phrase
•	procRun = signal logique “go” dans le pipe de sync (entre le même parent runc et son init).
•	exec.fifo = barrière inter-processus qui permet à une autre commande runc start (ou un orchestrateur) de décider quand débloquer l’init et lancer réellement le process.  ￼

Si tu veux, je peux te faire un mini diagramme séquence runc create/start montrant qui ouvre quoi et quand.