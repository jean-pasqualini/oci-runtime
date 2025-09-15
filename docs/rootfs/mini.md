mkdir -p /tmp/bundle/rootfs
cd /tmp/bundle
curl -fsSL \
https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/x86_64/alpine-minirootfs-latest-stable-x86_64.tar.gz \
-o alpine-minirootfs.tar.gz
tar -xpf alpine-minirootfs.tar.gz -C rootfs
cp /etc/resolv.conf rootfs/etc/resolv.conf

---

wget https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/aarch64/alpine-minirootfs-3.19.1-aarch64.tar.gz
tar -C $ROOT -xzf alpine-minirootfs-3.19.1-aarch64.tar.gz
rm alpine-minirootfs-3.19.1-aarch64.tar.gz