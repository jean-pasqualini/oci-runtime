Exact, tu as bien compris 👌

📌 Rôles séparés
•	containerd est responsable de créer le bundle OCI (donc rootfs/ + config.json).
•	Quand tu fais un ctr run ou via Kubernetes/CRI, containerd prépare le système de fichiers racine et génère le config.json conforme à la OCI runtime spec.
•	Ce fichier décrit exactement le process, les namespaces, les mounts, les cgroups, etc.
•	le shim n’a pas pour rôle de créer ce config.json.
•	Son rôle est de prendre ce bundle préparé et de traduire les appels de containerd (Create, Start, Kill, etc.) en commandes vers le runtime (ex: runc --bundle ... create <id>).
•	Autrement dit, il agit comme adapter / proxy entre containerd et le runtime choisi.

📌 Qui crée quoi ?
•	config.json → généré par containerd (ou par un outil en amont comme buildkit, nerdctl, ctr, cri-containerd via kubelet).
•	runc (ou autre runtime OCI) → lit ce config.json et applique ce qui est dedans pour mettre en place le conteneur.
•	shim → ne fait que relayer et adapter les appels, il ne fabrique pas la config.

⸻

👉 Donc oui : le config.json est créé en amont par containerd (ou par un composant qui l’utilise), pas par le shim.
Le shim sert uniquement d’intermédiaire runtime, il ne touche pas à la spec elle-même.

Veux-tu que je t’explique comment exactement containerd génère le config.json (et à partir de quoi, ex: image OCI, snapshotter, runtime options) ?