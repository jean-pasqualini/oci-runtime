Bonne question ⚡ — le champ Family de la struct unix.IfAddrmsg (ou struct ifaddrmsg côté kernel) indique de quelle famille d’adresses tu parles.

C’est exactement le même espace de valeurs que pour socket(AF_*, …) :

⸻

Valeurs principales que tu vas rencontrer

Constante Go (unix.*)	Valeur	Signification
AF_UNSPEC	0	Non spécifié (rare pour ifaddrmsg, sert souvent à dire “toutes familles”).
AF_INET	2	IPv4
AF_INET6	10	IPv6
AF_PACKET	17	Adresse link-layer (MAC par ex. pour sockaddr_ll). Utilisé surtout pour les sockets raw, pas pour RTM_NEWADDR.
AF_BRIDGE	7	Bridge interne du kernel. Rare.
AF_NETLINK	16	Famille Netlink elle-même.
AF_UNIX / AF_LOCAL	1	Adresses de type socket UNIX. Pas applicable à ifaddrmsg.


⸻

Dans le contexte de RTM_NEWADDR
•	Tu utilises presque toujours :
•	AF_INET → pour une adresse IPv4
•	AF_INET6 → pour une adresse IPv6
•	AF_UNSPEC peut être utilisé dans certaines requêtes (par ex. RTM_GETADDR) pour dire “je veux toutes les adresses peu importe la famille”.
•	Les autres (AF_PACKET, AF_BRIDGE, …) sont plutôt pour les sockets bas-niveau ou pour des netlink spécialisés, pas pour ajouter une IP à une interface.

⸻

Exemple pratique
•	Ajouter 192.168.1.10/24 sur eth0 :
→ Family = unix.AF_INET
•	Ajouter 2001:db8::5/64 sur eth0 :
→ Family = unix.AF_INET6
•	Lister toutes les adresses de toutes les familles :
→ Family = unix.AF_UNSPEC

⸻

👉 Donc pour ton runtime de conteneur, tu vas te limiter à AF_INET et AF_INET6, et peut-être AF_UNSPEC pour les requêtes génériques (genre dump d’adresses).

Veux-tu que je te fasse un tableau équivalent mais avec la vision conteneur (ex : lesquels sont utiles pour créer un container network stack, lesquels ne le sont pas) ?

-------


Exact 👌 — la liste des familles d’adresses (les AF_* constants) est très longue dans Linux, parce que chaque protocole ou sous-système du kernel peut avoir la sienne.
Dans Go, tu les trouves dans golang.org/x/sys/unix sous forme de constantes.

⸻

⚡ Panorama des AF_* (familles d’adresses) dans Linux

Voici les plus connus/communs :

Constante (unix.*)	Valeur (souvent)	Usage
AF_UNSPEC	0	Non spécifié (wildcard).
AF_UNIX / AF_LOCAL	1	Sockets Unix (IPC local).
AF_INET	2	IPv4.
AF_AX25	3	Amateur Radio AX.25.
AF_IPX	4	Novell IPX. (obsolète)
AF_APPLETALK	5	AppleTalk. (historique)
AF_NETROM	6	Amateur radio NET/ROM.
AF_BRIDGE	7	Bridge interne du kernel.
AF_ATMPVC	8	ATM PVCs.
AF_X25	9	X.25.
AF_INET6	10	IPv6.
AF_ROSE	11	Amateur radio ROSE.
AF_DECnet	12	DECnet.
AF_NETBEUI	13	NetBEUI. (rare)
AF_SECURITY	14	Usage sécurité (LSM).
AF_KEY	15	PF_KEY Key Management API.
AF_NETLINK	16	Netlink sockets (communication kernel/user).
AF_PACKET	17	Paquets bruts (niveau 2 Ethernet).
AF_ASH	18	Ash protocol.
AF_ECONET	19	Acorn Econet.
AF_ATMSVC	20	ATM SVCs.
AF_RDS	21	RDS sockets (Infiniband).
AF_SNA	22	IBM SNA.
AF_IRDA	23	Infrared Data Association.
AF_PPPOX	24	PPP over various transports.
AF_WANPIPE	25	WANPIPE protocol.
AF_LLC	26	Logical Link Control.
AF_IB	27	InfiniBand.
AF_MPLS	28	MPLS (Multiprotocol Label Switching).
AF_CAN	29	Controller Area Network (bus automobile/industriel).
AF_TIPC	30	Transparent Inter-Process Communication.
AF_BLUETOOTH	31	Bluetooth sockets.
AF_IUCV	32	IBM z/VM IUCV.
AF_RXRPC	33	RxRPC protocol.
AF_ISDN	34	Integrated Services Digital Network.
AF_PHONET	35	Nokia Phonet.
AF_IEEE802154	36	IEEE 802.15.4 (ZigBee, IoT).
AF_CAIF	37	Ericsson CAIF.
AF_ALG	38	Kernel crypto API (AF_ALG sockets).
AF_NFC	39	Near Field Communication.
AF_VSOCK	40	VSOCK (VM<->Host communication, ex: virtio-vsock).
AF_KCM	41	Kernel Connection Multiplexor.
AF_QIPCRTR	42	Qualcomm IPC Router.
AF_SMC	43	Shared Memory Communications (IBM SMC-R).
AF_XDP	44	XDP sockets (AF_XDP).


⸻

✅ Dans ton cas (runtime conteneur)

Tu vas pratiquement manipuler seulement :
•	AF_UNSPEC (requêtes génériques / dump)
•	AF_INET (IPv4)
•	AF_INET6 (IPv6)
•	(parfois) AF_PACKET si tu bosses avec veth et sockets raw
•	(optionnel) AF_NETLINK si tu parles netlink toi-même

Le reste (CAN, Bluetooth, NFC, VSOCK, XDP, etc.) est utile pour des applis très spécifiques (IoT, VM host comm, radio amateur, etc.).

⸻

👉 Veux-tu que je te prépare un tableau réduit “conteneur/networking only” avec juste les familles pertinentes à la construction d’un runtime de container ?