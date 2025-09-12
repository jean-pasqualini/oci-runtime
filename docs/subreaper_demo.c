// subreaper_demo.c
#define _GNU_SOURCE
#include <sys/prctl.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char **argv) {
    int enable_subreaper = 1;
    if (argc > 1 && argv[1][0] == '0') enable_subreaper = 0;

    pid_t me = getpid();
    printf("[subreaper] PID=%d  (argv[1]=%s => subreaper=%s)\n",
           me, argc>1?argv[1]:"(none)", enable_subreaper?"ON":"OFF");

    if (enable_subreaper) {
        if (prctl(PR_SET_CHILD_SUBREAPER, 1) != 0) {
            perror("prctl(PR_SET_CHILD_SUBREAPER)");
            return 1;
        }
    }

    pid_t child = fork();
    if (child < 0) { perror("fork child"); return 1; }

    if (child == 0) {
        // --- Processus parent intermédiaire (va mourir vite) ---
        pid_t mid = getpid();
        printf("[parent-int] PID=%d  PPID=%d -> fork du petit-fils\n", mid, getppid());

        pid_t grand = fork();
        if (grand < 0) { perror("fork grandchild"); exit(1); }

        if (grand == 0) {
            // --- Petit-fils (simule le PID 1 d’un conteneur) ---
            pid_t g = getpid();
            printf("[petit-fils] start  PID=%d  PPID=%d\n", g, getppid());
            // Attendre que le parent intermédiaire meure
            sleep(2);
            // Vérifier à qui on est rattaché après l’orphelinage
            printf("[petit-fils] après orphelinage  PID=%d  PPID=%d\n", g, getppid());
            // Laisser le temps au subreaper de faire wait()
            sleep(1);
            _exit(0);
        } else {
            // Le parent intermédiaire meurt rapidement -> orphelinage du petit-fils
            _exit(0);
        }
    } else {
        // --- Processus subreaper (ce programme) ---
        // Boucle de reap (wait) pour éviter les zombies
        int status;
        while (1) {
            pid_t w = waitpid(-1, &status, WNOHANG);
            if (w == 0) {           // personne à récolter pour l’instant
                usleep(100 * 1000);
            } else if (w > 0) {     // a récolté quelqu’un
                // printf("[subreaper] waitpid(%d) status=%d\n", w, status);
            } else {                // plus personne
                break;
            }
        }
    }
    return 0;
}