#!/bin/bash

set -e

check_prerequisites() {
	for c in git go goreleaser; do
		if ! [[ -x "$(command -v ${c})" ]]; then
			echo "ERR: missing command: '${c}'." >&2
			echo "Please install before continue." >&2
			exit 1
		fi
	done

	if [[ -z "$1" ]]; then
		echo "ERROR: give version number, e.g. v0.8.2"
		exit 1
	fi

	if [[ $1 =~ ^v[0-9]{1,2}\.[0-9]{1,2}\.[0-9]{1,3}$ ]]; then
		echo "version string format is CORRECT"
	else
		echo "version string format ins't correct"
		exit 1
	fi
exit 0

}

check_prerequisites $1

version="$1"
cur_branch=$(git rev-parse --abbrev-ref HEAD)
start_branch="master"

if [[ "$cur_branch" != "$start_branch" ]]; then
	echo "ERROR: checkout $start_branch branch"
	exit 1
fi

if [[ -z "$(git status --porcelain)" ]]; then
	git tag -a "$version" -m "v. $version"
	git push origin "$cur_branch" --tags
	GOPROXY=proxy.golang.org go list -m github.com/lainio/err2@"$version"
	goreleaser release --clean
else
	echo 'ERROR: working dir is not clean'
fi
