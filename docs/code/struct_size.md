Très bonne question 👍 En Go tu as plusieurs façons d’obtenir la taille d’une struct, mais il faut distinguer deux choses :

⸻

1. La taille en mémoire (avec padding)

👉 C’est ce que Go alloue réellement pour la struct, en respectant les règles d’alignement de l’architecture.

Tu utilises unsafe.Sizeof :

import "unsafe"

type Example struct {
    A uint8   // 1 octet
    B uint32  // 4 octets
}

fmt.Println(unsafe.Sizeof(Example{})) // → 8

Explication :
	•	A = 1 octet, mais B doit être aligné sur 4 octets → Go ajoute 3 octets de padding.
	•	Total = 1 + 3 (padding) + 4 = 8.

⸻

2. La taille “sans padding” pour un encodage binaire

👉 Quand tu fais binary.Write, il ne regarde pas le padding de Go, mais la taille logique des champs (enchaînés les uns après les autres).

Tu peux l’obtenir en additionnant les tailles fixes des champs avec unsafe.Sizeof champ par champ, ou plus simplement : en créant une instance, l’écrire avec binary.Write, et mesurer la longueur du buffer :

var buf bytes.Buffer
h := Example{}
_ = binary.Write(&buf, binary.LittleEndian, h)
fmt.Println(len(buf.Bytes())) // → 5 (1 + 4)


⸻

3. Pour les structs C exposées par Go (unix.IfInfomsg, etc.)

Si tu veux savoir exactement combien d’octets une struct C fait sur ta plateforme, tu passes aussi par unsafe.Sizeof :

import "golang.org/x/sys/unix"

fmt.Println(unsafe.Sizeof(unix.IfInfomsg{})) // souvent 16

👉 Ça reflète bien le layout utilisé par Go pour mapper la struct C.

⸻

4. Résumé pratique
	•	unsafe.Sizeof(struct{}) → taille réelle en mémoire avec padding Go.
	•	binary.Write → écrit la concaténation des champs, sans padding.
	•	Pour du Netlink : ce qui compte, c’est la taille C officielle (structs kernel), que tu retrouves souvent avec unsafe.Sizeof(unix.XXX{}).

⸻

⚠️ Question cruciale pour toi : tu veux la taille pour :
	•	savoir combien de bytes envoyer au kernel (== taille C exacte)
ou
	•	savoir combien de bytes binary.Write va générer (== somme des tailles des champs, pas forcément identique) ?