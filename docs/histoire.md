Voici une histoire (technique mais digeste) de l’« idée conteneur » sur Linux — d’où ça vient, comment ça marche sous le capot, et les jalons qui ont tout changé.

Les racines (avant Linux)
•	1979 – chroot (Unix V7) : première « prison » de processus en les enfermant dans un autre répertoire racine. Minimal, pas d’isolation réseau/CPU/mémoire.
•	2000 – FreeBSD Jails : isolation beaucoup plus complète (réseau, hostname, fs).
•	2004 – Solaris Zones : industrialise l’idée côté Sun.
Ces idées influenceront directement Linux.

Les briques du noyau Linux

Linux n’a pas « un » syscall magique pour les conteneurs : c’est un assemblage de mécanismes.
•	Namespaces (isolation des vues)
•	mnt (2002) : chaque conteneur voit son propre arbre de montage.
•	UTS & IPC (≈2006) : hostname/domain, files de messages/shm isolés.
•	PID (2008) : un PID 1 dans le conteneur, différent de l’hôte.
•	net (2008) : pile réseau dédiée (interfaces, routes, iptables).
•	user (2013) : mappe l’UID 0 dans le conteneur vers un utilisateur non-root sur l’hôte → base du rootless.
•	cgroups (contrôle/quotas)
Introduits fin 2007/début 2008 (2.6.24) pour limiter et mesurer CPU, mémoire, IO, etc.
cgroup v2 (2016+) unifie les contrôleurs et simplifie l’arborescence.
•	Sécurité
capabilities (droits fins au lieu du tout-puissant root), seccomp-BPF (filtre d’appels système), AppArmor/SELinux (MAC).
•	Systèmes de fichiers empilés
AUFS (hors mainline), overlayfs (dans le noyau 3.18, 2014) : permet d’assembler une image en couches (lecture) + un diff (écriture) → images légères et partageables.

Premiers « conteneurs Linux » (années 2000)
•	Linux-VServer (≈2001–2003) et OpenVZ (≈2005) offrent de la virtualisation OS-level avant la généralisation des namespaces complets dans mainline.
•	LXC (Linux Containers, ≈2008) : premier projet « pur Linux » qui assemble namespaces + cgroups pour fournir des conteneurs système. C’est bas niveau mais puissant.

2013 : Docker popularise l’approche
•	Docker (2013) part au début sur LXC, introduit la notion d’image en couches, un registre (Docker Hub) et surtout une DX (CLI, build via Dockerfile) qui rend le concept mainstream pour les devs.
•	libcontainer → runc (2014–2015) : Docker remplace LXC par sa propre librairie, puis donne runc (le binaire d’exécution bas niveau) à la communauté standardisée.

Standardisation : l’OCI
•	Open Container Initiative (OCI, 2015) : spécifications Image (format d’image) et Runtime (comment exécuter) → interopérabilité.
runc devient l’implémentation de référence ; d’autres runtimes apparaissent : crun (rapide, écrit en C), gVisor/runsc (sandbox user-space), Kata Containers (isolation via micro-VM).

Orchestrateurs et écosystème
•	Kubernetes (open-sourcé en 2014) change l’échelle : planifie des pods (groupes de conteneurs) avec un plan de contrôle, des Services, etc.
•	2016 : CRI (Container Runtime Interface) normalise le branchement des runtimes.
•	containerd (extrait de Docker, donné à la CNCF) devient un runtime de base courant.
•	CRIO/CRI-O (2016–2017) : runtime minimal compatible CRI pour Kubernetes.
•	2022 : suppression de dockershim dans K8s ; on utilise containerd/CRIO sous le capot.
•	Podman/Buildah/Skopeo (Red Hat, ≈2017+) : alternative « daemonless », rootless par défaut grâce aux user namespaces.
•	LXD (Ubuntu) : surcouche conviviale à LXC pour des conteneurs « système » (proches de VMs) avec images, profils, réseaux intégrés.

Sécurité & isolation renforcées
•	Rootless (2019+) : exécuter des conteneurs sans droits root sur l’hôte devient courant (Podman, Docker rootless).
•	Sandboxing : gVisor (2018, interception syscalls en user-space), Kata (VMs ultra-légères), Firecracker (2018, micro-VMs, très utilisé pour FaaS) lorsque le besoin d’isolation est fort (multi-tenant).

Comment ça marche (en résumé)
1.	Le runtime (ex. runc, crun) crée des namespaces pour le processus (PID/net/mnt/uts/ipc/user).
2.	Il attache le processus à des cgroups qui imposent des limites CPU/mémoire/IO.
3.	Il monte un rootfs (souvent overlayfs) constitué des couches d’une image OCI + un « upperdir » en écriture.
4.	Il réduit les privilèges : capabilities minimales, profil seccomp, AppArmor/SELinux.
5.	Le processus démarre comme PID 1 dans son espace, avec son propre hostname, sa pile réseau (veth/bridge) et sa vue du système.

Pourquoi le raz-de-marée après 2013 ?
•	Expérience développeur : empaqueter « tout ce qu’il faut pour exécuter l’app » (binaire + libs + conf) → moins d’« ça marche sur ma machine ».
•	Images en couches : rapides à distribuer, cacheables, reproductibles (CI/CD).
•	Standardisation (OCI) + orchestration (K8s) : du laptop au cluster sans changer le modèle.
•	Efficacité : pas d’hyperviseur complet → densité et démarrage ultra-rapides (ms–s).

Aujourd’hui, les grands axes
•	cgroup v2 partout et tooling mature.
•	Rootless par défaut de plus en plus fréquent.
•	Runtimes spécialisés selon le besoin : performance (crun), compat (runc), iso renforcée (Kata/gVisor).
•	Chaîne d’approvisionnement : signatures (cosign), SBOM (SPDX/CycloneDX), images minimalistes (« distroless »), politiques d’admission (OPA/Gatekeeper/Kyverno).
•	Stockage : overlay2 dominant, alternatives (btrfs, zfs) selon les cas.

Repères chronologiques (TL;DR)
•	1979 : chroot (Unix).
•	2000–2004 : Jails/Zones (influence).
•	2002–2013 : namespaces Linux (mnt, uts/ipc, pid/net, user).
•	2007/2008 : cgroups v1.
•	2008 : LXC.
•	2013 : Docker.
•	2014 : overlayfs mainline.
•	2015 : OCI (Image/Runtime), runc.
•	2016+ : cgroup v2, CRI (K8s).
•	2017 : containerd → CNCF, Podman/Buildah.
•	2018 : gVisor, Kata, Firecracker.
•	2022 : Kubernetes retire dockershim (containerd/CRIO).

Si tu veux, je peux compléter avec un schéma des namespaces, une démo montrant comment créer un mini-conteneur à la main avec unshare/chroot, ou une comparaison Docker vs Podman vs LXC/LXD selon tes usages