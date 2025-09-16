En bref : c’est normal. 🙂

Dans config.json de runc, /etc/resolv.conf est bien un bind-mount depuis un fichier du host (/var/lib/docker/containers/<id>/resolv.conf).
Mais la commande mount (format “legacy”) n’affiche pas la source du bind ; elle affiche le périphérique du système de fichiers qui contient cette source.
Ici, ce fichier vit sur la partition ext4 du host montée depuis /dev/vda1, donc tu vois :

/dev/vda1 on /etc/resolv.conf type ext4 (...)

Pour voir la vraie origine d’un bind-mount, il faut regarder les infos détaillées du kernel :
•	cat /proc/self/mountinfo | grep " /etc/resolv.conf "
→ tu y verras deux champs clefs :
•	le device (ex. 253:1 → /dev/vda1)
•	le champ root (le chemin dans ce FS) qui pointe vers l’original (p.ex. /var/lib/docker/containers/<id>/resolv.conf).

Commandes utiles (plus lisibles) :

findmnt -T /etc/resolv.conf   # montre SOURCE/TARGET et indique "bind"
findmnt -no SOURCE,TARGET,PROPAGATION -T /etc/resolv.conf

Quelques précisions :
•	type ext4 reflète le FS d’origine du bind, pas “bind” (qui est un flag de montage).
•	Les options rbind + rprivate dans config.json disent à runc de faire un bind récursif et d’isoler la propagation des montages.

Donc rien d’anormal : tu es bien sur un bind-mount, et /dev/vda1 n’est que le device sous-jacent du fichier source sur le host.