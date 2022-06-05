#!/bin/bash

location=$(dirname "$BASH_SOURCE")

. "$location"/functions.sh

set -e

start_branch=$(git rev-parse --abbrev-ref HEAD)
migration_branch=${1:-"err2-auto-update"}
# TODO: remove
no_build_check=${no_build_check:-""}
use_current_branch=${use_current_branch:-""}

if [[ ! -z $use_current_branch ]]; then
	migration_branch="$start_branch"
fi

# =================== main =====================
# print_env

check_prerequisites

echo "update err2 package to latest version"
setup_repo
deps
check_build
commit "commit deps"

echo "====== basic err2 refactoring starts now ===="
replace_easy1

replace_1
replace_2

add_try_import
goimports_to_changed

# test and commit
check_build_and_pick
# commit "phase 1"

echo "====== complex refactoring starts now ===="
multiline_3
check_build_and_pick
# check_build
# commit "phase 2 multilines"

multiline_2
check_build_and_pick
# check_build
# commit "phase 2 multilines"

multiline_1
# check_build
# commit "phase 2 multilines"

goimports_to_changed
check_build
commit "phase 2 multilines"

