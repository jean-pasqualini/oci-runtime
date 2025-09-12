Parfait 👌 allons dans le détail de cette ligne, parce que c’est une des subtilités clés quand on lit plusieurs messages Netlink dans un seul recv.

⸻

1. Rappel : un recvfrom peut renvoyer plusieurs messages Netlink d’un coup

Le protocole Netlink envoie un “paquet” qui peut contenir un ou plusieurs messages Netlink consécutifs :

[ NetlinkMsg1 ][ NetlinkMsg2 ][ NetlinkMsg3 ] ...

Chaque message a une taille donnée par h.Len.
Mais ces tailles doivent être arrondies à 4 octets (alignement) pour que le suivant commence correctement.

⸻

2. Pourquoi l’alignement ?

La spec Netlink impose que chaque message Netlink commence à une adresse multiple de 4.
Ça permet de :
•	garantir que les structures internes (u32, u64) sont alignées en mémoire,
•	simplifier le parsing en C (pas de “misaligned access” sur certaines archis).

Donc après un message de h.Len octets, il faut avancer au multiple de 4 supérieur.

⸻

3. Comment est calculé adv

Le code :

adv := (int(h.Len) + unix.NLMSG_ALIGNTO - 1) & ^(unix.NLMSG_ALIGNTO - 1)

Avec NLMSG_ALIGNTO = 4.

C’est une astuce binaire pour dire :
👉 “arrondis h.Len au multiple de 4 supérieur”

Exemple :
•	h.Len = 16 → (16 + 3) & ^3 = 16 (pas de padding)
•	h.Len = 18 → (18 + 3) & ^3 = 20 (on saute 2 octets de padding)
•	h.Len = 21 → (21 + 3) & ^3 = 24

⸻

4. Avancer p

p = p[adv:]

Ça veut dire :

“découpe le slice p et avance de adv octets, pour que p pointe maintenant sur le prochain message”.

Si on avait reçu 3 messages d’un coup, la boucle va s’exécuter ainsi :

Itération 1: p pointe sur Msg1 → traiter Msg1 → couper p = p[adv1:]
Itération 2: p pointe sur Msg2 → traiter Msg2 → couper p = p[adv2:]
Itération 3: p pointe sur Msg3 → traiter Msg3 → couper p = p[adv3:]

Quand len(p) devient < NLMSG_HDRLEN, on arrête : il n’y a plus de message complet.

⸻

5. Illustration visuelle

Mémoire reçue (p initial) :

+-----------------------+----------------------+------------------+
| NlMsghdr | payload... | NlMsghdr | payload.. | NlMsghdr | ...   |
+-----------------------+----------------------+------------------+
^
p

Après traitement du premier message (adv = ALIGN(h.Len)):

+-----------------------+----------------------+------------------+
|  Msg1 (adv bytes)     | NlMsghdr | payload.. | NlMsghdr | ...   |
+-----------------------+----------------------+------------------+
^
p


⸻

6. Que se passerait-il sans alignement ?

Si tu faisais naïvement :

p = p[h.Len:]

→ tu pourrais tomber au milieu d’un padding plutôt qu’au début du message suivant.
Résultat : (*NlMsghdr)(unsafe.Pointer(&p[0])) lirait des octets corrompus → parsing cassé, erreurs type bad nlmsg len.

⸻

✅ En résumé :
•	adv = h.Len arrondi à 4 octets.
•	Ça garantit que tu tombes toujours pile sur l’entête du message suivant.
•	p = p[adv:] permet de parcourir un buffer contenant plusieurs messages Netlink d’affilée, de manière sûre.

⸻

Veux-tu que je t’écrive une petite fonction Go align4(n int) int avec des exemples concrets de tailles (13, 16, 21, …) pour voir ce que ça donne ?

-----

Parfait 🙌
Voici une petite fonction utilitaire align4 qui fait exactement le même calcul que dans ton code (NLMSG_ALIGN) et quelques exemples pour visualiser :

package main

import "fmt"

// arrondit n au multiple de 4 supérieur
func align4(n int) int {
return (n + 4 - 1) & ^(4 - 1)
}

func main() {
vals := []int{13, 16, 17, 18, 21, 32}
for _, v := range vals {
fmt.Printf("n=%2d → align4(n)=%2d\n", v, align4(v))
}
}

Résultat :

n=13 → align4(n)=16
n=16 → align4(n)=16
n=17 → align4(n)=20
n=18 → align4(n)=20
n=21 → align4(n)=24
n=32 → align4(n)=32


⸻

Comment ça marche
•	n + 3 : on s’assure de dépasser le multiple de 4 si n n’est pas déjà aligné.
•	& ^3 : masque les deux derniers bits (équivalent à arrondir vers le bas à un multiple de 4).
•	Donc ça revient à dire : “donne-moi le plus petit multiple de 4 ≥ n”.

⸻

Pourquoi 4 ?

Parce que le protocole Netlink impose un alignement à 4 octets (NLMSG_ALIGNTO = 4).
Si tu travaillais avec des nla (attributs), c’est la même logique : chaque attribut est aussi aligné à 4.

⸻

👉 Veux-tu que je pousse l’exemple en montrant un buffer contenant deux messages (par ex. un message de 13 octets et un de 21 octets) et comment align4 permet de retrouver le deuxième proprement ?