#!/bin/bash

set -e
location=$(dirname "$BASH_SOURCE")

filtered_build() {
	local osname=$(uname -s)
	local pkg=${1:-"./..."}
	local awk_file="$location"/delete-"$osname".awk

	del=$(go build -o /dev/null "$pkg" 2>&1 >/dev/null | grep 'declared but not used' | awk -F : -f "$awk_file") 

	if [[ $del != "" ]]; then
		eval $del
		echo "FILTER"
	else
		echo "OK"
	fi
}

pkg=${1:-"./..."}

result=$(filtered_build "$pkg")

echo $result

