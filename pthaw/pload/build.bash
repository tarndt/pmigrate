#!/bin/bash
set -eu

echo -n "Compiling ploader program... "
trap 'echo "Failed."' EXIT
gcc -std=gnu99 -static -fno-stack-protector -nostartfiles -nostdlib -ffreestanding -c *.c
echo "OK"

echo -n "Linking ploader program... "
trap 'echo "Failed."' EXIT
ld.bfd -m elf_x86_64 -static -z max-page-size=0x1000 --defsym RESERVE_TOP=0 --script pload.ld *.o -o ploader
echo "OK"

trap '' EXIT
rm *.o
exit
