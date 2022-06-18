#!/bin/bash

location=$(dirname "$BASH_SOURCE")
. $location/functions.sh

set -e

# =================== main =====================
while getopts 'dnvoum:' OPTION; do
	case "$OPTION" in
	n)
		vlog "no commits"
		no_commit=1
		;;
	d)
		echo "set verbose/debug mode"
		verbose=1
		;;
	v)
		echo "set verbose/debug mode"
		verbose=1
		;;
	o)
		vlog "running only simple migrations"
		only_simple=1
		;;
	u)
		vlog "using current branch"
		use_current_branch=1
		;;
	m)
		migration_branch="$OPTARG"
		vlog "migration_branch = $OPTARG"
		;;
	?)
		echo "usage: $(basename $0) [-n] [-v] [-o] [-u] [-m runmode] [functions...]" >&2
		echo "       n: no commit" >&2
		echo "       d: add debug output" >&2
		echo "       v: verbose" >&2
		echo "       o: only simple migrations" >&2
		echo "       u: using current branch" >&2
		echo "       m: migration branch" >&2
		exit 1
		;;
	esac
done
shift "$(($OPTIND -1))"

migration_branch=${migration_branch:-"err2-auto-update"}
no_commit=${no_commit:-"1"}
start_branch=$(git rev-parse --abbrev-ref HEAD)
use_current_branch=${use_current_branch:-""}
only_simple=${only_simple:-""}

if [[ ! -z $use_current_branch ]]; then
	vlog "owerride migration branch with current branch"
	vlog "use_current_branch: $use_current_branch"
	migration_branch="$start_branch"
fi

vlog "location: $location"
vlog "$BASH_SOURCE"

for a in "$@"; do
	$a
done
