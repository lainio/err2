#!/bin/bash

set -e
location=$(dirname "$BASH_SOURCE")
osname=$(uname -s)
awk_file="$location"/delete-"$osname".awk

del=$(go build -o /dev/null ./... 2>&1 >/dev/null | grep 'declared but not used' | awk -F : -f "$awk_file") 

if [[ $del != "" ]]; then
	eval $del
else
	echo "OK"
fi
