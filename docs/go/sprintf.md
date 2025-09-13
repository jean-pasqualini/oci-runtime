Parfait 👍 Voici un cheatsheet fmt.Printf en Go avec les verbes de formatage les plus utiles. Tu pourras l’utiliser comme référence rapide :

⸻

🔤 Chaînes de caractères
•	%s → chaîne brute
•	%q → chaîne entre guillemets, caractères spéciaux échappés
•	%x → chaîne en hexadécimal (deux chiffres par byte)
•	%X → idem mais en majuscules
•	%#U → caractère Unicode (U+1234 '⌂')

⸻

🔢 Entiers
•	%d → base 10
•	%b → binaire
•	%o → octal
•	%x / %X → hexadécimal
•	%c → caractère (interprété comme rune)
•	%q → caractère entre quotes (avec échappement si nécessaire)

⸻

🔢 Nombres flottants & complexes
•	%f → décimal (par défaut, 6 décimales)
•	%.2f → 2 décimales (arrondi)
•	%e / %E → notation scientifique (expo en min/maj)
•	%g / %G → choisit le plus compact (%f ou %e)
•	%b → représentation technique (mantisse/exposant)

⸻

✅ Booléens
•	%t → true ou false

⸻

📦 Pointeurs
•	%p → adresse mémoire (en hexadécimal)
•	%#p → adresse avec préfixe 0x

⸻

📑 Structs, slices, maps, interfaces
•	%v → valeur “par défaut” (format humain lisible)
•	%+v → inclut les noms de champs des structs
•	%#v → syntaxe Go (rejouable en code Go si possible)
•	%T → type de la valeur
•	%t → booléen
•	%p → pointeur

⸻

⚙️ Divers
•	%% → littéral %
•	%n → écrit dans un int passé par pointeur le nombre de bytes imprimés

⸻

📐 Largeur et alignement
•	%6d → largeur minimale 6 (aligné à droite)
•	%-6d → largeur minimale 6 (aligné à gauche)
•	%06d → largeur 6, padding avec 0

⸻

🚀 Exemples pratiques

fmt.Printf("%s\n", "hello")     // hello
fmt.Printf("%q\n", "hello\n")   // "hello\n"

fmt.Printf("%d\n", 42)          // 42
fmt.Printf("%04d\n", 42)        // 0042

fmt.Printf("%f\n", 3.14159)     // 3.141590
fmt.Printf("%.2f\n", 3.14159)   // 3.14
fmt.Printf("%e\n", 3.14159)     // 3.141590e+00

fmt.Printf("%v\n", []int{1,2})  // [1 2]
fmt.Printf("%#v\n", []int{1,2}) // []int{1, 2}

type P struct {X, Y int}
fmt.Printf("%+v\n", P{1, 2})    // {X:1 Y:2}

fmt.Printf("%T\n", 42)          // int


⸻

👉 Tu veux que je t’en fasse une tableau résumé imprimable (genre markdown bien compact) que tu peux garder comme antisèche ?