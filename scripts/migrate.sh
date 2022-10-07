#!/bin/bash

location=$(dirname "$BASH_SOURCE")

. "$location"/functions.sh

set -e

# =================== main =====================
while getopts 'dvoushm:' OPTION; do
	case "$OPTION" in
	d)
		echo "set verbose/debug mode"
		debug=1
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
		avalue="$OPTARG"
		vlog "The runmode provided is $OPTARG"
		;;
	s)
		allow_subdir=1
		vlog "Allowing subdir processing"
		;;
	h|?)
		echo "usage: $(basename $0) [-d] [-v] [-o] [-u] [-m runmode] [migration_branch]" >&2
		echo "       d: add debug output" >&2
		echo "       v: verbose" >&2
		echo "       o: only simple migrations" >&2
		echo "       u: using current branch" >&2
		echo "       s: allow subdir processing" >&2
		echo "       m: reserved" >&2
		exit 1
		;;
	esac
done
shift "$(($OPTIND -1))"

migration_branch=${1:-"err2-auto-update"}

start_branch=$(git rev-parse --abbrev-ref HEAD)
use_current_branch=${use_current_branch:-""}
only_simple=${only_simple:-""}

if [[ ! -z $use_current_branch ]]; then
	vlog "owerride migration branch with current branch"
	vlog "use_current_branch: $use_current_branch"
	migration_branch="$start_branch"
fi

print_env

check_prerequisites

vlog "update err2 package to latest version"
setup_repo
deps
check_build ./...
commit "commit deps"

echo "====== basic err2 refactoring ===="
echo "processing..."

replace_annotate
replace_easy1
replace_2

add_try_import
goimports_to_changed

bads=$(check_build_and_pick)

check_if_stop_for_simplex

echo "====== complex refactoring 1. ===="

try_0
try_3
try_2
try_1

bads=$(check_build_and_pick)

echo "====== complex refactoring 2. ===="

multiline_0
bads=$(check_build_and_pick)

multiline_3
bads=$(check_build_and_pick)

multiline_2
bads=$(check_build_and_pick)

multiline_1
bads=$(check_build_and_pick)

echo
echo "================================="
echo "====== auto-refactoring done ===="
echo "================================="
echo

if [[ $bads != "" ]]; then
	echo "====== TODO Summary ====" >&2
	echo "Please check the following files before commit:" >&2
	echo "" >&2
	echo "$bads" | tr " " "\n" >&2
	exit 1
fi

echo "Easy peasy! All done."

