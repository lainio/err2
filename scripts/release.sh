#!/bin/bash

set -e

if [[ -z "$1" ]]; then
	echo "ERROR: give version number, e.g. v0.8.2"
	exit 1
fi

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
else
	echo 'ERROR: working dir is not clean'
fi
