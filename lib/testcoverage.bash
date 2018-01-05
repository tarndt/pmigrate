#!/bin/bash

PACKAGES=$(shell find ./ -type d -not -path '*/\.*')

echo "mode: count" > coverage-all.out
$(foreach pkg,$(PACKAGES),\
	go test -coverprofile=coverage.out -covermode=count $(pkg);\
	tail -n +2 coverage.out >> coverage-all.out;)
go tool cover -html=coverage-all.out


#!/bin/bash

echo "mode: set" > acc.out
for Dir in $(find ./* -maxdepth 10 -type d );
do
        if ls $Dir/*.go &> /dev/null;
        then
            go test -coverprofile=profile.out $Dir
            if [ -f profile.out ]
            then
                cat profile.out | grep -v "mode: set" >> acc.out
            fi
fi
done
go tool cover -html=acc.out
rm -rf ./profile.out
rm -rf ./acc.out
