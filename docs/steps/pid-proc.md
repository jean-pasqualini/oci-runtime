```
	unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, "")
	unix.Unmount("/proc", unix.MNT_DETACH)
	unix.Mount("proc", "/proc", "proc", uintptr(unix.MS_NOSUID|unix.MS_NODEV|unix.MS_NOEXEC), "")

```

Top—voilà ce que font, précisément, ces trois lignes et pourquoi l’ordre compte :

unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, "")

	•	But : casser la propagation des montages héritée de l’hôte dans ton mount namespace.
	•	Comment : avec source="" et fstype="", mount(2) n’effectue pas un nouveau montage ; il modifie les propriétés de propagation du point de montage cible.
	•	Flags :
	•	MS_PRIVATE : rend le(s) montage(s) privés → aucune propagation ni entrante, ni sortante.
	•	MS_REC : applique récursivement à / et tous ses sous-montages.
	•	Pourquoi ici ? Sur la plupart des systèmes (systemd), / est rshared. Si tu ne fais pas ça, tes (dé)montages peuvent se propager vers l’hôte.
	•	⚠️ À faire dans un mount namespace neuf (CLONE_NEWNS/unshare -m), sinon tu affectes l’hôte.

⸻


unix.Unmount("/proc", unix.MNT_DETACH)

	•	But : détacher le /proc hérité (celui de l’hôte) de ton mount ns.
	•	MNT_DETACH (aka lazy unmount) : détache immédiatement le point de montage de l’arborescence du namespace, même s’il est “occupé” ; le démontage réel se fera quand les dernières références disparaîtront.
	•	Pratique et inoffensif ici : si /proc n’était pas monté, l’appel peut retourner une erreur — qu’on ignore souvent sciemment.

⸻


unix.Mount("proc", "/proc", "proc", uintptr(unix.MS_NOSUID|unix.MS_NODEV|unix.MS_NOEXEC), "")

	•	But : monter un procfs neuf qui reflète le PID namespace courant.
	•	source="proc", fstype="proc" : c’est la manière standard de monter procfs.
	•	Options de sécurité :
	•	MS_NOSUID : ignore les bits setuid/setgid.
	•	MS_NODEV  : pas de périphériques.
	•	MS_NOEXEC : on n’exécute rien depuis /proc.
	•	Effet voulu : si ton process est dans un nouveau namespace PID (CLONE_NEWPID), ce /proc ne montrera que les PIDs de ce ns → ps ne voit plus l’hôte.

⸻

À retenir (pièges & bonnes pratiques)
•	Ordre : MS_PRIVATE (récursif) → detacher l’ancien /proc → monter le nouveau /proc.
•	Capacités : ces mount nécessitent CAP_SYS_ADMIN dans le mount ns. En non-root, crée un user namespace (NEWUSER + UID/GID mappings) pour l’obtenir localement.
•	Créer le point de montage : assure-toi que /proc existe :

_ = os.MkdirAll("/proc", 0555)


	•	Optionnel : tu peux ajouter des montages “classiques” d’un conteneur :
	•	/sys en RO : sysfs + MS_RDONLY|MS_NOSUID|MS_NODEV|MS_NOEXEC
	•	/dev en tmpfs (mode 755), devpts pour les TTY, /dev/shm (tmpfs 1777)
	•	hidepid : si tu veux restreindre l’accès entre utilisateurs dans le même ns, procfs supporte data="hidepid=2" (option data du mount). Ce n’est pas requis pour l’isolation entre ns.

En résumé : la 1ʳᵉ ligne isole la propagation, la 2ᵉ retire l’ancien /proc, la 3ᵉ installe un /proc propre à ton namespace → condition nécessaire pour que ps ne voie que “son monde”.