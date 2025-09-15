#include <stdio.h>

int main() {
    char couleur[100];

    printf("Quelle est ta couleur favorite ? ");
    fflush(stdout); // force l'affichage immédiat du prompt

    if (fgets(couleur, sizeof(couleur), stdin) != NULL) {
        // Supprimer le saut de ligne à la fin, s'il existe
        char *newline = couleur;
        while (*newline != '\0') {
            if (*newline == '\n') {
                *newline = '\0';
                break;
            }
            newline++;
        }

        printf("Tu as choisi : %s\n", couleur);
    } else {
        printf("\nErreur de lecture.\n");
    }

    return 0;
}