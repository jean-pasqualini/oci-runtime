Pour être efficace plutôt qu'exploratoire, voici ce qui me serait utile, par ordre de priorité :

1. Un handler existant complet (ex. create ou start) — me donne d'un coup : signature, parsing des flags urfave/cli/v3, gestion d'erreur, logging, structure générale.
2. Le point de câblage dans cmd/oci-runtime/ — comment les sous-commandes sont enregistrées (root command, liste des commands, etc.).
3. Le modèle d'état persisté sous --root — quelle structure de fichiers/dossiers représente un container (status, pid, bundle path…). C'est essentiel pour savoir ce que delete doit nettoyer côté state.
4. Le contrat IPC — s'il y a un process container vivant, comment on lui parle (signal ? message ? socket ?) pour l'arrêter avant cleanup.
5. L'API de infrastructure/linux/ — la liste des fonctions de nettoyage disponibles (unmount, supprimer ns, cgroups…) plutôt que de deviner.
6. Les conventions transverses — pattern d'erreur dans infrastructure/technical/error, logger, et la forme du modèle Container dans internal/domain/.
7. La sémantique OCI attendue — delete doit-il échouer si le container tourne ? Y a-t-il un --force ? Quels états du cycle de vie sont valides ?

Si tu me donnes (1) + (2) + (3) en markdown ou collé directement, je peux écrire le squelette du handler sans rien lire. Les points 4–7 ne deviennent nécessaires qu'au moment d'implémenter le corps réel du nettoyage.

✻ Crunched for 17s

❯ tu est d'accord que le problème d'un agent peu performant c'est quand il doit faire du discovery à chaque fois

Oui, complètement. Le coût réel n'est pas la découverte elle-même (inévitable la première fois) mais sa répétition à chaque session : relire les mêmes fichiers, redéduire les mêmes conventions, reposer les mêmes questions. Le levier, c'est de capitaliser ce qui a été découvert dans des supports durables — CLAUDE.md (déjà bien fait ici avec la project map), des templates de handlers, voire des mémoires — pour que le prochain agent démarre informé plutôt qu'aveugle.

✻ Baked for 8s