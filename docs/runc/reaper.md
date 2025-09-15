https://github.com/opencontainers/runc/blob/55c90aaf023331bd6f7fca9c000899cf27b2686f/utils_linux.go#L221C1-L308C2
https://github.com/opencontainers/runc/blob/55c90aaf023331bd6f7fca9c000899cf27b2686f/signals.go#L127
---

Oui — ton intuition est bonne : runc run peut se comporter comme “child subreaper” pour le conteneur, afin de ré-associer et reaper les processus orphelins (zombies) quand l’init du conteneur fait un double-fork/daemonize. Regarde les points clés du code que tu as collé :
•	handlerCh := newSignalHandler(r.enableSubreaper, r.notifySocket)
→ si r.enableSubreaper est vrai (piloté par le flag CLI --subreaper, généralement activé par défaut), newSignalHandler fait un prctl(PR_SET_CHILD_SUBREAPER, 1).
→ Effet : runc devient subreaper pour l’arbre de processus du conteneur, et reçoit les SIGCHLD des petits-enfants, ce qui lui permet de les wait()/reap proprement.
•	switch r.action … CT_ACT_RUN: err = r.container.Run(process)
→ CT_ACT_RUN fait bien create + start en un seul appel.
•	tty, err := setupIO(...) puis tty.waitConsole() / tty.ClosePostStart()
→ gestion I/O et console (si --tty, via --console-socket).
•	handler := <-handlerCh puis status, err := handler.forward(process, tty, detach)
→ la boucle de gestion des signaux (SIGCHLD, SIGINT, SIGTERM, SIGWINCH, etc.) s’occupe de :
1.	attendre la fin du process init,
2.	récupérer et répercuter les statuts d’exit,
3.	reaper les enfants si subreaper est actif.
•	detach := r.detach || (r.action == CT_ACT_CREATE)
→ en mode détaché (ou en create), runc ne reste pas en avant-plan ; sinon il reste parent direct de l’init (cas run non détaché) et gère les signaux jusqu’à l’exit de l’init.
•	defer func(){ if err != nil { r.destroy() } } et plus bas if err == nil { r.destroy() }
→ housekeeping : nettoyer le bundle/état à la fin si on n’est pas en mode détaché.

Conclusion rapide
•	runc run (non détaché) : runc lance l’init et reste au premier plan, et se pose en subreaper (si option activée) pour récu­pérer les zombies et gérer les signaux.
•	Pas de shim : comme il n’y a pas de containerd-shim, c’est justement ce mode subreaper qui évite les zombies lorsque l’init daemonize.
•	create + start séparés : même logique de subreaper côté handler si --subreaper est activé, mais runc start sort rapidement après le lancement (le petit process runc:[1:INIT] reste parent dans ce mode).

Si tu veux, je peux t’indiquer la portion de code dans newSignalHandler où le PR_SET_CHILD_SUBREAPER est fait et où les SIGCHLD sont traités.