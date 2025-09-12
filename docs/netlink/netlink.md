github.com/vishvananda/netlink
Top — voici un mini-kit Netlink en Go (sans lib tierce) pour faire les opérations de base : UP une interface, ajouter une IP, ajouter une route par défaut.
Ça utilise uniquement golang.org/x/sys/unix (+ stdlib).

⚠️ Nécessite CAP_NET_ADMIN dans le netns courant.
Code Linux-only, simplifié pour rester lisible (pas de gestion avancée des erreurs/alignements multiples messages).

⸻

1) Mettre une interface UP (RTM_NEWLINK)

//go:build linux

package netutil

import (
"encoding/binary"
"fmt"
"net"
"unsafe"

	"golang.org/x/sys/unix"
)

func BringUp(name string) error {
ifi, err := net.InterfaceByName(name)
if err != nil { return fmt.Errorf("iface %q: %w", name, err) }

	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil { return fmt.Errorf("netlink socket: %w", err) }
	defer unix.Close(fd)

	const IFF_UP = 0x1
	ifm := unix.IfInfomsg{
		Family: unix.AF_UNSPEC,
		Index:  int32(ifi.Index),
		Flags:  IFF_UP,
		Change: IFF_UP,
	}

	msg := nlMsg(unix.RTM_NEWLINK, unix.NLM_F_REQUEST|unix.NLM_F_ACK, ifinfomsgBytes(ifm))
	if err := nlSend(fd, msg); err != nil { return err }
	return nlRecvAck(fd)
}

func ifinfomsgBytes(m unix.IfInfomsg) []byte {
// sizeof(struct ifinfomsg) = 16
b := make([]byte, 16)
b[0] = byte(m.Family)
// b[1] pad, b[2..3] type=0
binary.LittleEndian.PutUint32(b[4:], uint32(m.Index))
binary.LittleEndian.PutUint32(b[8:], uint32(m.Flags))
binary.LittleEndian.PutUint32(b[12:], uint32(m.Change))
return b
}


⸻

2) Ajouter une adresse IP (RTM_NEWADDR)

func AddrAdd(name, cidr string) error {
ifi, err := net.InterfaceByName(name)
if err != nil { return fmt.Errorf("iface %q: %w", name, err) }

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil { return fmt.Errorf("parse CIDR %q: %w", cidr, err) }
	fam := unix.AF_INET
	if ip.To4() == nil { fam = unix.AF_INET6 }

	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil { return fmt.Errorf("netlink socket: %w", err) }
	defer unix.Close(fd)

	// ifaddrmsg
	// struct ifaddrmsg { u8 family; u8 prefixlen; u8 flags; u8 scope; u32 index; }
	ia := make([]byte, 8)
	ia[0] = byte(fam)
	ones, _ := ipnet.Mask.Size()
	ia[1] = byte(ones)  // prefixlen
	ia[2] = 0           // flags
	ia[3] = unix.RT_SCOPE_UNIVERSE
	binary.LittleEndian.PutUint32(ia[4:], uint32(ifi.Index))

	// Attributes: IFA_LOCAL + IFA_ADDRESS (pour IPv4 mettre les deux identiques)
	attrs := nlAttrs()
	ipb := ip.To4()
	if ipb == nil { ipb = ip.To16() }
	attrs.Add(unix.IFA_LOCAL,  ipb)
	attrs.Add(unix.IFA_ADDRESS, ipb)

	payload := append(ia, attrs.Bytes()...)
	msg := nlMsg(unix.RTM_NEWADDR, unix.NLM_F_REQUEST|unix.NLM_F_ACK|unix.NLM_F_CREATE|unix.NLM_F_EXCL, payload)

	if err := nlSend(fd, msg); err != nil { return err }
	return nlRecvAck(fd)
}


⸻

3) Ajouter une route par défaut (RTM_NEWROUTE)

func RouteAddDefault(name, gw string) error {
ifi, err := net.InterfaceByName(name)
if err != nil { return fmt.Errorf("iface %q: %w", name, err) }

	ip := net.ParseIP(gw)
	if ip == nil { return fmt.Errorf("bad gw %q", gw) }
	fam := unix.AF_INET
	if ip.To4() == nil { fam = unix.AF_INET6 }

	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil { return fmt.Errorf("netlink socket: %w", err) }
	defer unix.Close(fd)

	// rtmsg: family, dst_len=0 (default), src_len=0, tos=0, table=RT_TABLE_MAIN, proto=RTPROT_BOOT, scope=RT_SCOPE_UNIVERSE, type=RTN_UNICAST
	rt := make([]byte, 12)
	rt[0] = byte(fam)                       // family
	// rt[1]=dst_len 0, rt[2]=src_len 0, rt[3]=tos 0
	rt[4] = unix.RT_TABLE_MAIN              // table
	rt[5] = unix.RTPROT_BOOT                // proto
	rt[6] = unix.RT_SCOPE_UNIVERSE
	rt[7] = unix.RTN_UNICAST
	// rt[8..11] flags = 0

	attrs := nlAttrs()
	gwb := ip.To4()
	if gwb == nil { gwb = ip.To16() }
	attrs.Add(unix.RTA_GATEWAY, gwb)
	attrs.Add(unix.RTA_OIF, u32(uint32(ifi.Index)))

	payload := append(rt, attrs.Bytes()...)
	msg := nlMsg(unix.RTM_NEWROUTE, unix.NLM_F_REQUEST|unix.NLM_F_ACK|unix.NLM_F_CREATE|unix.NLM_F_EXCL, payload)

	if err := nlSend(fd, msg); err != nil { return err }
	return nlRecvAck(fd)
}


⸻

4) Petites primitives Netlink (header, attrs, send/ack)

// Common helpers (put in the same package)

func nlMsg(typ, flags int, payload []byte) []byte {
hlen := unix.NLMSG_HDRLEN
b := make([]byte, hlen+len(payload))
h := (*unix.NlMsghdr)(unsafe.Pointer(&b[0]))
h.Len = uint32(len(b))
h.Type = uint16(typ)
h.Flags = uint16(flags)
h.Seq = 1
copy(b[hlen:], payload)
return b
}

type nlattrBuf struct{ b []byte }

func nlAttrs() *nlattrBuf { return &nlattrBuf{} }

func (a *nlattrBuf) Add(typ int, value []byte) {
const align = unix.NLA_ALIGNTO
l := unix.NLA_HDRLEN + len(value)
pad := (l + align - 1) & ^(align - 1)
nb := make([]byte, pad)
binary.LittleEndian.PutUint16(nb[0:2], uint16(l))
binary.LittleEndian.PutUint16(nb[2:4], uint16(typ))
copy(nb[unix.NLA_HDRLEN:], value)
a.b = append(a.b, nb...)
}

func (a *nlattrBuf) Bytes() []byte { return a.b }

func u32(v uint32) []byte {
var b [4]byte
binary.LittleEndian.PutUint32(b[:], v)
return b[:]
}

func nlSend(fd int, msg []byte) error {
return unix.Sendto(fd, msg, 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK})
}

func nlRecvAck(fd int) error {
buf := make([]byte, 8192)
n, _, err := unix.Recvfrom(fd, buf, 0)
if err != nil { return fmt.Errorf("recv: %w", err) }
p := buf[:n]
for len(p) >= unix.NLMSG_HDRLEN {
h := (*unix.NlMsghdr)(unsafe.Pointer(&p[0]))
if int(h.Len) < unix.NLMSG_HDRLEN || int(h.Len) > len(p) {
return fmt.Errorf("bad nlmsg len")
}
msg := p[unix.NLMSG_HDRLEN:int(h.Len)]
if h.Type == unix.NLMSG_ERROR {
if len(msg) < 4 { return fmt.Errorf("short NLMSG_ERROR") }
code := int32(binary.LittleEndian.Uint32(msg[:4]))
if code == 0 { return nil } // ACK OK
return fmt.Errorf("netlink error: %d", code)
}
// next (4-byte aligned)
adv := (int(h.Len) + unix.NLMSG_ALIGNTO - 1) & ^(unix.NLMSG_ALIGNTO - 1)
if adv <= 0 || adv > len(p) { break }
p = p[adv:]
}
return nil
}


⸻

Exemple d’utilisation (dans ton init)

// Assure lo UP, donne une IP à eth0, ajoute la route par défaut
if err := netutil.BringUp("lo"); err != nil { return err }
if err := netutil.BringUp("eth0"); err != nil { return err }
if err := netutil.AddrAdd("eth0", "172.17.0.3/16"); err != nil { return err }
if err := netutil.RouteAddDefault("eth0", "172.17.0.1"); err != nil { return err }


⸻

Conseils & pièges
•	Capas : CAP_NET_ADMIN requis; dans Docker: --cap-add NET_ADMIN (ou --privileged).
•	Netns : fais ça dans le netns cible (ou setns() avant).
•	IPv6 : ajuste family et longueurs (le code ci-dessus le fait).
•	Erreurs Netlink : si netlink error: -17 → EEXIST (existe déjà).
•	Production : gère les multi-messages, timeouts, et écoute des événements (RTM_NEWLINK, RTM_DELADDR) si besoin.

Si tu veux, je peux te rajouter une fonction AddrDel (RTM_DELADDR) et RouteDelDefault, ou un monitor qui log les évènements link up/down via RTNLGRP_LINK.

----


Parfait 👍 Je vais te faire un résumé structuré du protocole Netlink pour tes notes, avec un mélange de contexte, de fonctionnement interne et d’exemples.

⸻

📘 Netlink : protocole noyau ↔ espace utilisateur

1. Définition
   •	Netlink est un protocole de communication bidirectionnel entre le noyau Linux et les processus en espace utilisateur.
   •	Utilisé pour la configuration et la supervision de nombreux sous-systèmes du noyau : réseau, pare-feu, cgroups, audit, etc.
   •	Implémenté comme une famille de sockets (AF_NETLINK).

⸻

2. Objectifs
   •	Remplacer les anciens ioctl réseau (SIOCSIFFLAGS, SIOCSIFADDR, …).
   •	Fournir une API :
   •	Extensible (nouveaux types de messages facilement ajoutables).
   •	Structurée (messages typés avec attributs).
   •	Événementielle (le noyau peut notifier les applis).

⸻

3. Architecture

Côté user-space :
•	Ouvre un socket Netlink :

socket(AF_NETLINK, SOCK_RAW, NETLINK_ROUTE);


	•	Construit un message Netlink (NlMsghdr + payload spécifique).
	•	Envoie au noyau (sendto).
	•	Reçoit les réponses/événements (recvfrom).

Côté kernel :
•	Chaque famille Netlink traite certains types de messages.
•	Exemples :
•	NETLINK_ROUTE → interfaces, adresses, routes.
•	NETLINK_NETFILTER → iptables/nftables.
•	NETLINK_GENERIC → framework extensible (ex: WireGuard, nl80211 pour le WiFi).

⸻

4. Structure des messages

Un message Netlink a deux couches :
1.	En-tête Netlink (NlMsghdr)
•	len : taille totale du message.
•	type : type du message (ex: RTM_NEWLINK, RTM_NEWADDR).
•	flags : options (NLM_F_REQUEST, NLM_F_ACK, NLM_F_CREATE…).
•	seq : numéro de séquence.
•	pid : identifiant de l’émetteur (0 = noyau).
2.	Payload (varie selon la famille)
•	Ex: IfInfomsg pour configurer une interface, IfAddrmsg pour ajouter une adresse, Rtmsg pour les routes.
•	Suivi d’une liste d’attributs (NlAttr), type/longueur/valeur (TLV).

⸻

5. Exemple : mettre lo UP
    1.	User crée un message RTM_NEWLINK :
          •	Header : NlMsghdr{Type=RTM_NEWLINK, Flags=NLM_F_REQUEST|NLM_F_ACK}
          •	Payload : IfInfomsg{Index=1, Flags=IFF_UP, Change=IFF_UP}
    2.	Envoi via sendto().
    3.	Kernel reçoit, modifie les flags de lo.
    4.	Kernel répond avec un NLMSG_ERROR (code=0) → ACK.

⸻

6. Mode événement
   •	User peut rester en recvfrom() pour écouter les notifications :
   •	RTM_NEWLINK quand une interface monte ou descend.
   •	RTM_NEWADDR quand une IP est ajoutée.
   •	RTM_DELROUTE quand une route disparaît.
   •	C’est ce que fait ip monitor link.

⸻

7. Usages typiques
   •	Réseau (NETLINK_ROUTE) :
   •	RTM_NEWLINK / RTM_DELLINK → interfaces.
   •	RTM_NEWADDR / RTM_DELADDR → adresses IP.
   •	RTM_NEWROUTE / RTM_DELROUTE → routes.
   •	Pare-feu (NETLINK_NETFILTER) : nftables, conntrack.
   •	Wireless (GENL_ID_CTRL via Generic Netlink) : WiFi (nl80211).
   •	Sécurité : audit, SELinux.

⸻

8. Avantages par rapport à ioctl
   •	Extensible : pas besoin d’ajouter un nouvel ioctl pour chaque feature.
   •	Lisible/structuré : attributs TLV au lieu de structs figées.
   •	Événementiel : le noyau peut pousser des infos (ex: lien Ethernet down).

⸻

9. Limitations
   •	Assez bas niveau → manipulation fastidieuse en C/Go.
   •	Pas stable en ABI : il faut inclure les en-têtes du noyau (linux/rtnetlink.h).
   •	Plus complexe que ioctl pour des cas simples.

⸻

10. Résumé
    •	Netlink = protocole noyau ↔ user extensible via sockets.
    •	Base de la commande ip et des libs comme vishvananda/netlink.
    •	Introduit dans Linux 2.2, stabilisé avec Generic Netlink en 2.6.
    •	Aujourd’hui incontournable pour toute configuration réseau avancée.

⸻

👉 Veux-tu que je t’ajoute un diagramme visuel (ASCII ou image) qui montre le chemin ip link set lo up → netlink socket → kernel → ACK ?