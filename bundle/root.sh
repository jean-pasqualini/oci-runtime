#!/bin/bash
set -e

ROOT=/app/bundle/root
rm -rf $ROOT
mkdir -p $ROOT/{bin,lib,lib/aarch64-linux-gnu}

cp -av /bin/bash $ROOT/bin/

#cp -a linux-vdso.so.1 $ROOT
cp -av /lib/aarch64-linux-gnu/libtinfo.so.6 $ROOT/lib/aarch64-linux-gnu/
cp -av /lib/aarch64-linux-gnu/libc.so.6 $ROOT/lib/aarch64-linux-gnu/
cp -av /lib/ld-linux-aarch64.so.1 $ROOT/lib/

cp -av /lib/aarch64-linux-gnu/libtinfo.so.6.5 $ROOT/lib/aarch64-linux-gnu/
cp -av /lib/aarch64-linux-gnu/ld-linux-aarch64.so.1 $ROOT/lib/aarch64-linux-gnu/

# 3) dépendances
#ldd /bin/bash | awk '/=> \//{print $3} /^\/lib/{print $1}' | xargs -I{} cp -a {} "$ROOT"/$(dirname {} | sed "s#^/##")/

# Copier les dépendances de bash
#ldd /bin/bash | awk '/=> \// {print $3} /^\/lib/ {print $1}' | while read lib; do
#  mkdir -p $ROOT/$(dirname $lib)
#   echo "cp -a $lib $ROOT$lib"
#  cp -a $lib $ROOT$lib
#done