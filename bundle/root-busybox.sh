#!/bin/bash
set -e

ROOT=/app/bundle/root
rm -rf $ROOT
mkdir -p $ROOT

wget https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/aarch64/alpine-minirootfs-3.19.1-aarch64.tar.gz
tar -C $ROOT -xzf alpine-minirootfs-3.19.1-aarch64.tar.gz
rm alpine-minirootfs-3.19.1-aarch64.tar.gz