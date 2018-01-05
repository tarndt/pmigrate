#!/bin/bash
set -eu

echo "Building C programs statically in: `pwd`..."

for FILE in *.c; do	
	NAME="`basename $FILE`"
	NAME="${NAME%.*}"
	echo -en "\t* Building: $NAME...\t"
	trap "echo FAILURE for: $NAME" EXIT
	gcc -std=c99 -static $FILE -o $NAME
	echo "OK"
done

trap "" EXIT
echo "All programs built."

exit
