Exactement ✅

containerd est conçu pour être agnostique du runtime OCI concret. C’est pour ça qu’il expose plusieurs shims, chacun servant de “colle” entre containerd et un runtime particulier.

⸻

🔹 Comment ça marche
•	Un shim est un petit binaire qui implémente une API gRPC standardisée par containerd.
•	Il est chargé dynamiquement selon le runtime_type configuré dans config.toml.
•	Chaque shim connaît les détails d’appel de son runtime associé (flags, conventions, gestion des états…).

⸻

🔹 Exemples de shims existants
•	containerd-shim-runc-v2 → wrapper officiel pour runc (par défaut).
•	containerd-shim-crun-v2 → wrapper pour crun.
•	containerd-shim-kata-v2 → pour Kata Containers (VM isolées).
•	containerd-shim-gvisor-v2 → pour gVisor (runtime user-space).
•	etc.

⸻

🔹 Choix du shim

Dans /etc/containerd/config.toml tu peux définir plusieurs runtimes :

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
runtime_type = "io.containerd.runc.v2"

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata]
runtime_type = "io.containerd.kata.v2"

Et côté Kubernetes, dans le RuntimeClass CRD, tu peux choisir quel runtime (donc quel shim) utiliser par pod.

⸻

👉 Donc oui : containerd met à dispo plusieurs shims (chacun implémentant la même API vers containerd, mais parlant un “dialecte” différent vers son runtime OCI).

Veux-tu que je te montre visuellement le schéma des couches (containerd → shim → runtime → kernel) pour que ça soit encore plus clair ?