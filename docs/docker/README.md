Bonne question 👌
Docker n’a pas disparu avec containerd, mais son rôle a changé : aujourd’hui il délègue énormément, mais pas tout.

⸻

🔹 Architecture actuelle (Docker “moderne”)

Quand tu lances docker run ... :
1.	Docker CLI
•	C’est juste le binaire docker que tu tapes dans ton shell.
•	Il envoie une requête HTTP à dockerd (le démon Docker).
2.	dockerd (Docker Engine)
•	C’est lui le vrai orchestrateur “Docker”.
•	Il expose l’API REST de Docker.
•	Depuis Docker 1.11 (2016), dockerd ne gère plus directement les conteneurs.
•	Il délègue à containerd la gestion des images, snapshots, et conteneurs.
3.	containerd (intégré dans Docker)
•	Reçoit les ordres de dockerd.
•	Prépare le bundle OCI (rootfs + config.json).
•	Gère les snapshots (via overlayfs, btrfs, zfs…).
•	Lance un shim v2 adapté au runtime choisi.
4.	shim v2 (ex: containerd-shim-runc-v2)
•	Traduit l’API Task de containerd en appels concrets au runtime.
5.	OCI Runtime (par défaut runc)
•	Lit le config.json
•	Configure namespaces, cgroups, mounts, etc.
•	Lance le process du conteneur.

⸻

🔹 Donc, Docker fait quoi exactement ?
•	Docker CLI : ergonomie (UX, options conviviales).
•	dockerd :
•	API REST + authentification,
•	gestion des volumes, réseaux, plugins,
•	intégration Swarm, secrets, etc.,
•	délégation à containerd pour tout ce qui est bas-niveau.
•	containerd :
•	moteur de conteneurs conforme OCI,
•	gère images, snapshots, bundle, shims.
•	runc :
•	runtime OCI qui applique le config.json.

⸻

👉 Résumé :
•	Avant 2016 : Docker gérait directement tout (image → conteneur via libcontainer → runc).
•	Depuis : Docker s’appuie sur containerd pour l’exécution et la gestion bas-niveau.
•	Docker reste surtout une surcouche orientée développeurs (CLI, API REST, UX), alors que containerd est le runtime “pur” utilisé aussi par Kubernetes.

⸻

Veux-tu que je te fasse un schéma complet du pipeline Docker → containerd → shim → runc pour visualiser qui fait quoi ?

------

Tu as raison : ce n’est pas une bonne pratique de “mettre tout le runtime” dans le shim.
Le shim doit rester mince : un adaptateur qui implémente l’API Task (v2) de containerd et la traduit vers ton runtime.

Pourquoi éviter un “gros shim”
•	Séparation des responsabilités :
•	containerd : images/snapshots/bundles + orchestration.
•	shim : RPC Task (Create/Start/Exec/Kill/Wait/Delete), gestion du stdio, propagation d’événements, reaper.
•	runtime : application du config.json (namespaces, cgroups, mounts…), exécution.
•	Stabilité & maintenance : un shim gras devient un mini-containerd + mini-runtime à maintenir (cgroups v1/v2, seccomp, apparmor, idmaps, mounts…).
•	Isolation des pannes : un shim simple est plus robuste et remplaçable sans impacter le reste.
•	Évolutivité : si ton runtime évolue, un shim fin limite la surface à mettre à jour.

Modèles qui marchent bien
1.	Fork/exec d’un runtime OCI (runc, crun)
•	Le shim appelle un binaire runtime existant (… create/start/…) et relaie l’état.
•	Le plus simple et le plus courant.
2.	Daemon/runtime “sandbox” séparé (Kata, gVisor)
•	Le shim parle à TON démon/agent (VM, user-space kernel, etc.) et reste un pont TTRPC.
•	Le “gros” de la logique vit dans le runtime/agent, pas dans le shim.
3.	Librairie runtime embarquée (linker une lib type libcontainer)
•	Possible, mais attention : tu ré-implémentes alors toute la logique OCI dans le shim → à éviter sauf raison impérieuse.

Ce que le shim devrait faire
•	Implémenter l’API Task v2 (TTRPC) : Create, Start, Kill, Exec, Wait, Delete, State, Events, Checkpoint/Restore (si support).
•	Gérer le lifecycle des processus, le reaping, le stdio, les PIDs/exit codes.
•	Traduire les appels vers ton runtime (CLI, socket, gRPC interne, agent VM, etc.).
•	Ne pas : créer config.json, gérer les images/snapshots, ni réécrire la spec OCI.

Quand “grossir” un peu le shim ?
•	Prototypage rapide.
•	Besoin minimal de colle spécifique (ex : bootstrap d’un agent/sandbox).
Même là, vise une fine couche et déplace le cœur métier dans un binaire/runtime dédié.

Reco pratique (structure de projet)
•	shim/ : serveur TTRPC + mapping des RPC Task → appels runtime (mince).
•	runtime/ : vrai moteur (daemon/CLI/lib) qui applique la spec (lourd).
•	Tests d’intégration via ctr/nerdctl avec un runtime_type = "io.containerd.tonruntime.v2".

En bref : garde le shim petit, mets l’intelligence dans le runtime. C’est plus sûr, plus maintenable, et ça colle à l’architecture containerd.

----

Yes 👍, le shim s’appuie sur le layout d’état que maintient le runtime (ex. runc) sous le --root qu’on lui passe.

⸻

🔹 Concrètement avec containerd-shim-runc-v2

Quand containerd demande au shim :

runc --root /run/containerd/runc --bundle ... create <id>

	•	runc crée un dossier d’état :

/run/containerd/runc/<id>/
├─ state.json      # état courant du conteneur
├─ pid            # PID du process init
├─ attach         # fifos pour stdio
└─ ...


	•	Ce layout contient tout ce que le runtime doit connaître pour manipuler le conteneur après le create.
	•	Quand le shim reçoit un State RPC de containerd, il appelle en réalité :

runc --root /run/containerd/runc state <id>

et parse le JSON renvoyé (qui est basé sur les fichiers dans ce répertoire d’état).

	•	Pour un Kill, Exec, Delete, etc., le shim passe aussi par ce répertoire (--root) pour retrouver le conteneur et agir dessus.

⸻

🔹 Donc : qui consulte quoi ?
•	Le runtime (runc) : lit et écrit dans le layout (state.json, pid, fifos…).
•	Le shim : ne manipule pas directement les fichiers, mais il s’appuie sur les commandes CLI (state, kill, delete, …) qui elles-mêmes utilisent ce layout.
•	containerd : ne touche pas au layout d’état → il demande tout au shim, qui s’appuie sur le runtime.

⸻

👉 En résumé :
•	Oui, le shim consulte indirectement le layout d’état via les commandes du runtime.
•	C’est le runtime (runc/crun/…) qui définit et maintient ce layout.
•	Le shim, lui, se contente de relayer les appels (il ne devrait pas réimplémenter la logique d’état).

Veux-tu que je te montre à quoi ressemble exactement un state.json de runc et quelles infos le shim en retire pour répondre à containerd ?

----


Exactement ✅, tu as bien résumé :
•	Le shim v2 ne connaît pas les détails internes du runtime (layout d’état, fichiers temporaires, etc.).
•	Sa seule dépendance, c’est la façon dont il dialogue avec le runtime :
•	soit via une CLI (comme runc, où il appelle runc state, runc kill, etc.),
•	soit via une API interne (socket RPC, gRPC, agent dans une VM comme Kata/gVisor),
•	soit via une lib si on l’embarque (plus rare, mais possible).

Donc, si ton runtime expose la même interface CLI que runc, tu peux réutiliser le shim existant (containerd-shim-runc-v2).
Sinon, tu fais ton shim spécifique, mais il reste fin : juste un traducteur entre l’API Task de containerd et l’API/CLI de ton runtime.

👉 En résumé : le shim n’est pas couplé au layout d’état, seulement au contrat d’interface de ton runtime.

Veux-tu que je te prépare une checklist des appels CLI/outputs que ton runtime doit émuler pour être 100% compatible avec containerd-shim-runc-v2 sans recoder de shim ?

-----

Bonne question 👌

📌 containerd expose deux types d’API principales :
1.	gRPC API (la principale)
•	Containerd expose un daemon (containerd) qui ouvre une socket Unix (par défaut /run/containerd/containerd.sock).
•	Les clients comme ctr, nerdctl, ou le plugin CRI de Kubernetes parlent à containerd via gRPC sur cette socket.
•	L’API gRPC est très large : gestion d’images, snapshots, conteneurs, tâches, événements, etc.
•	La spec des services est dans containerd/api (fichiers .proto).
2.	CRI plugin (gRPC aussi, pour Kubernetes)
•	Kubernetes n’appelle pas directement l’API native de containerd.
•	kubelet parle au plugin CRI de containerd via gRPC (toujours sur une socket Unix, ex /run/containerd/containerd.sock).
•	Ce plugin traduit les appels CRI (CreateContainer, StartContainer, etc.) en appels containerd internes.

📌 Pas d’API HTTP/REST
•	Containerd n’expose pas de REST HTTP (contrairement à Docker Engine qui expose une API HTTP/REST).
•	Tout passe par gRPC (souvent avec TTRPC côté interne — une version plus légère pour les shims et process internes).

⸻

👉 Donc :
•	Externe (clients/CRI) → gRPC sur Unix socket.
•	Interne (containerd ↔ shim) → TTRPC (plus léger que gRPC).

⸻

Veux-tu que je te montre un exemple concret de commande grpcurl que tu peux lancer sur /run/containerd/containerd.sock pour explorer l’API gRPC de containerd ?

-----

Exactement ✅
On peut dire qu’il y a deux couches de RPC autour de containerd, mais elles ne sont pas identiques :

⸻

🔹 1. containerd ↔ clients (docker, nerdctl, kubelet/CRI, ctr…)
•	Protocole : gRPC “classique” (Protobuf, HTTP/2, socket Unix).
•	C’est l’API publique de containerd.
•	Elle couvre : images, snapshots, namespaces, tâches, événements, etc.
•	Exemple de service : containerd.services.containers.v1.Containers.
•	C’est via ça que kubelet parle au plugin CRI de containerd, ou que ctr/nerdctl fonctionnent.

⸻

🔹 2. containerd ↔ shim v2 (runc, crun, kata, gVisor…)
•	Protocole : TTRPC (Tiny Transport RPC), une version plus légère que gRPC (même IDL, mais optimisée).
•	C’est l’API interne définie dans api/runtime/task/v2/task.proto.
•	Là on a le fameux service Task avec Create, Start, Kill, Exec, State, etc.
•	Chaque shim implémente ce service et expose une socket (/run/containerd/.../shim.sock).
•	Containerd envoie ses requêtes au shim via TTRPC.

⸻

🔹 Pourquoi deux couches ?
•	gRPC externe → API riche, extensible, standard, pour outils & orchestrateurs.
•	TTRPC interne → beaucoup plus léger (pas de HTTP/2, pas de TLS, moins de copies mémoire), adapté au hot path des conteneurs et aux nombreux shims qu’il faut gérer.

⸻

👉 Donc :
•	Oui, il y a bien deux niveaux :
•	gRPC pour “monter” (clients → containerd),
•	TTRPC pour “descendre” (containerd → shims).

⸻

Veux-tu que je te dessine un schéma ASCII clair avec ces deux niveaux (clients → gRPC → containerd → TTRPC → shim → runtime) ?

---

Bonne précision 👌

👉 Entre containerd et les shims v2, ce n’est pas gRPC mais TTRPC.

⸻

📌 Pourquoi TTRPC et pas gRPC ?
•	TTRPC (Tiny Transport RPC) est une implémentation minimaliste, faite par les mainteneurs de containerd.
•	Elle garde les .proto (donc mêmes définitions qu’en gRPC) mais supprime tout l’empilement HTTP/2, TLS, etc.
•	Résultat :
•	beaucoup plus rapide (moins d’allocations, moins de copies mémoire),
•	binaire plus petit,
•	idéal pour la comm locale via une socket Unix entre containerd et le shim.

⸻

📌 Donc la stack est :
•	Clients externes → containerd : gRPC “normal” (sur /run/containerd/containerd.sock).
•	containerd → shim v2 : TTRPC (sur une socket dédiée du shim, ex /run/containerd/io.containerd.runtime.v2.task/.../shim.sock).

⸻

👉 En résumé :
•	gRPC = API publique de containerd (pour kubelet, ctr, nerdctl, etc.).
•	TTRPC = API interne entre containerd et les shims.

⸻

Veux-tu que je te montre un petit extrait d’un .proto (ex. task.proto) et son équivalent TTRPC, pour voir comment ça se traduit concrètement ?

------

Bonne question 👌

Le nom TTRPC veut dire Tiny Transport RPC.
Et justement, son transport est très simple :

📌 Transport TTRPC
•	Socket Unix (le plus courant)
→ chaque shim v2 ouvre une socket locale (shim.sock), et containerd s’y connecte.
•	Ou fichier/pipe (possible mais rare)
→ techniquement, TTRPC fonctionne tant qu’on a un FD type stream.

En clair, TTRPC se limite à un flux de bytes point-à-point (pas de HTTP/2, pas de framing complexe).

⸻

📌 Comment ça marche
•	Les .proto définissent les services (ex : Create, Start, State du service Task).
•	Un générateur TTRPC produit du code Go ou Rust avec :
•	un codec binaire très proche de Protobuf,
•	un framing minimal : chaque message = header (taille, type, ID) + payload protobuf.
•	Containerd ouvre la socket du shim et envoie directement les appels RPC encodés → le shim répond de la même manière.

⸻

📌 Différence avec gRPC
•	gRPC = Protobuf + HTTP/2 (streaming, multiplexage, TLS, metadata).
•	TTRPC = Protobuf + framing minimal + transport raw (socket Unix).
•	👉 Résultat : beaucoup plus léger et rapide, mais limité à de la comm locale.

⸻

👉 Donc : le transport TTRPC, c’est une socket Unix locale (raw stream), avec un protocole maison simplifié basé sur Protobuf.

Veux-tu que je te décrive le framing d’un message TTRPC (header + payload) pour voir concrètement ce qui circule entre containerd et un shim ?

----


Bonne question 👌

En Go, il existe une implémentation officielle de TTRPC (par les mainteneurs de containerd) et quelques extensions.

⸻

📌 Librairies Go pour TTRPC
1.	github.com/containerd/ttrpc
•	📦 lib principale et officielle.
•	Fournit le client/serveur TTRPC, génération de stubs, gestion du framing et de la socket.
•	C’est celle utilisée dans containerd et containerd-shim-*.
2.	github.com/containerd/ttrpc/proto
•	Outils pour générer du code Go TTRPC à partir de .proto.
•	S’utilise avec protoc + plugin protoc-gen-ttrpc.
3.	github.com/containerd/ttrpc/encoding
•	Support de l’encodage Protobuf, hooks pour en définir d’autres si besoin.

⸻

📌 Workflow typique en Go
1.	Tu écris ton fichier .proto (exemple simplifié) :

syntax = "proto3";

package example.v1;

service Echo {
rpc SayHello(HelloRequest) returns (HelloReply);
}

message HelloRequest {
string name = 1;
}

message HelloReply {
string message = 1;
}


	2.	Tu génères le code Go TTRPC :

protoc --go_out=. --ttrpc_out=. api.proto


	3.	Tu obtiens :
	•	api.ttrpc.pb.go → interfaces et stubs TTRPC.
	•	api.pb.go → structs Protobuf classiques.
	4.	Dans ton shim ou serveur Go, tu implémentes l’interface générée et tu démarres un serveur TTRPC sur une socket Unix.

⸻

👉 Donc, pour faire du TTRPC en Go, tu utilises principalement :
•	containerd/ttrpc (lib de runtime),
•	protoc-gen-ttrpc (plugin de génération).

⸻

Veux-tu que je te prépare un exemple Go minimal (serveur + client TTRPC qui parlent via une socket Unix) pour voir ça en pratique ?