Bonne question 👌

👉 Pour mettre une interface UP (avec netlink ou même ip link set dev X up) :
•	Tu n’es pas obligé d’ajouter une IP.
•	Exemple : ip link set lo up ne met aucune IP nouvelle, mais active l’interface.
•	Pour lo, le kernel crée déjà 127.0.0.1/8 (et ::1/128 si IPv6 activé), même si l’interface est DOWN. Ces adresses sont là par défaut.
•	Donc “mettre UP” suffit à les rendre utilisables.
•	Tu n’es pas obligé d’ajouter une route.
•	Pour lo, la route 127.0.0.0/8 dev lo est déjà dans la table, mais inactive tant que lo est DOWN.
•	Une fois lo UP, la route est utilisable immédiatement.

⸻

⚡ Résumé
•	lo a déjà une IP (127.0.0.1) et une route associée.
•	Tu dois juste l’amener UP (IFF_UP).
•	Pas besoin d’ajouter quoi que ce soit pour que ping 127.0.0.1 marche.

⸻

Exemple (équivalent Go → netlink)

if err := netutil.BringUp("lo"); err != nil {
return err
}

Exemple (shell)

ip link set lo up
# pas besoin de "ip addr add 127.0.0.1/8 dev lo"


⸻

✅ Donc pour ton use case (init minimal du conteneur) : un BringUp("lo") suffit, pas besoin d’IP ni de route.

Veux-tu que je t’explique pourquoi malgré ça certaines libs/runtimes réappliquent quand même 127.0.0.1/8 sur lo (genre containerd, runc) ?

-----


Bonne remarque 👌

Tu as bien vu : github.com/vishvananda/netlink est très haut-niveau. Il expose des fonctions pratiques (LinkSetUp, AddrAdd, RouteAdd, …) mais sous le capot il fait beaucoup de choses (décodage d’attributs, cache des links, compatibilité IPv6, etc.).

⸻

⚙️ Les couches possibles pour “parler Netlink”
1.	Très bas niveau : golang.org/x/sys/unix
•	Tu fais toi-même : socket(AF_NETLINK, …), sendmsg, recvmsg.
•	Tu construis à la main NlMsghdr, IfInfomsg, RtAttr en bytes.
•	➡️ C’est ce qu’on a commencé à écrire ensemble (RTM_NEWLINK à la main).
•	Ultra flexible, mais beaucoup de boilerplate.

⸻

	2.	Niveau intermédiaire (juste le protocole Netlink)
	•	Exemple : github.com/mdlayher/netlink
	•	Fournit une API Go propre pour ouvrir un socket netlink et envoyer/recevoir des NlMsghdr.
	•	Mais ne connaît pas RTM_NEWLINK/RTM_NEWADDR/RTM_NEWROUTE → à toi de parser les payloads.
	•	Tu écris ton propre binding pour IfInfomsg, RtAttr, etc.
	•	➡️ C’est le “milieu” que tu cherches : plus agréable que x/sys/unix, mais pas aussi riche que vishvananda/netlink.

⸻

	3.	Haut niveau (use-cases réseau complets)
	•	github.com/vishvananda/netlink
	•	Connaît déjà toutes les structures (IfInfomsg, RtMsg, IfAddrmsg, …).
	•	Expose directement LinkSetUp, LinkByName, AddrAdd, RouteAdd, etc.
	•	Utilisé par runc, containerd, CNI plugins.
	•	➡️ Très pratique, mais lourd si tu veux juste manipuler un flag.

⸻

✅ Résumé
•	Bas niveau : x/sys/unix → tout à la main.
•	Intermédiaire : mdlayher/netlink → gestion propre des sockets Netlink, mais payloads à parser soi-même.
•	Haut niveau : vishvananda/netlink → tout prêt, complet (mais plus lourd).

⸻

👉 Veux-tu que je te montre un exemple concret d’utilisation de mdlayher/netlink pour faire un RTM_NEWLINK (mettre lo UP), histoire de voir à quoi ressemble cette couche “intermédiaire” ?