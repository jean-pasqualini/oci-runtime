#!/bin/bash
dir=$(dirname $0)
gcc -static -o $dir/read $dir/read.c