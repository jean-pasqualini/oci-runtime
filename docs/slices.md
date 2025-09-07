Yes 👌 bonne question.

Depuis Go 1.21, le langage a introduit un nouveau package standard : slices (et son cousin maps).
C’est un ajout majeur car jusque-là, Go n’avait pas beaucoup d’outils génériques pour manipuler les slices/arrays.

⸻

🚀 Nouvelles fonctions de manipulation de slices/arrays (Go 1.21+)

🔹 Recherche / comparaison
•	slices.Equal(a, b) → compare deux slices élément par élément.
•	slices.EqualFunc(a, b, eq func(T, T) bool) → version custom avec ton comparateur.
•	slices.Compare(a, b) → compare lexicographiquement (-1, 0, 1).
•	slices.CompareFunc(a, b, cmp func(T, T) int) → version custom.

🔹 Recherche d’éléments
•	slices.Index(s, v) → retourne l’index du premier v, ou -1.
•	slices.IndexFunc(s, func(T) bool) → index du premier élément qui match la condition.
•	slices.Contains(s, v) → bool.
•	slices.ContainsFunc(s, func(T) bool) → bool.

🔹 Copie et extraction
•	slices.Clone(s) → retourne une copie indépendante du slice.
•	slices.Delete(s, i, j) → supprime les éléments [i:j].
•	slices.DeleteFunc(s, func(T) bool) → supprime tous les éléments qui matchent.
•	slices.Insert(s, i, vals...) → insère à la position i.
•	slices.Replace(s, i, j, vals...) → remplace [i:j] par de nouveaux éléments.

🔹 Tri et recherche triée
•	slices.Sort(s) → trie selon l’ordre naturel (<).
•	slices.SortFunc(s, cmp func(T, T) int) → tri custom.
•	slices.BinarySearch(s, v) → recherche dichotomique dans un slice trié.
•	slices.BinarySearchFunc(s, x, cmp func(T, T) int) → version custom.

🔹 Divers
•	slices.Compact(s) → supprime les doublons consécutifs.
•	slices.CompactFunc(s, eq func(T, T) bool) → version custom.
•	slices.Reverse(s) → inverse l’ordre.
•	slices.IsSorted(s) / slices.IsSortedFunc(s, cmp) → vérifie si c’est trié.
•	slices.Min(s) / slices.Max(s) → plus petit / plus grand élément.
•	slices.MinFunc(s, cmp) / slices.MaxFunc(s, cmp) → versions custom.

⸻

⚡ Exemple rapide

package main

import (
"fmt"
"slices"
)

func main() {
nums := []int{5, 3, 8, 3, 1}

	slices.Sort(nums)        // tri
	fmt.Println(nums)        // [1 3 3 5 8]

	nums = slices.Compact(nums)
	fmt.Println(nums)        // [1 3 5 8]

	i, found := slices.BinarySearch(nums, 5)
	fmt.Println(i, found)    // 2 true

	nums = slices.Insert(nums, 1, 42)
	fmt.Println(nums)        // [1 42 3 5 8]
}


⸻

👉 Donc, si tu viens du monde PHP/JS où tu as array_map, array_filter, etc., maintenant Go a un équivalent beaucoup plus riche via slices (et maps pour les dictionnaires).

Tu veux que je te fasse un tableau comparatif entre les fonctions array_* de PHP et leurs équivalents slices.* en Go pour que tu voies direct les parallèles ?