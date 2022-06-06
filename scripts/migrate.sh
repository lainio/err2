#!/bin/bash

location=$(dirname "$BASH_SOURCE")

. "$location"/functions.sh

set -e

start_branch=$(git rev-parse --abbrev-ref HEAD)
migration_branch=${1:-"err2-auto-update"}
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

echo "====== basic err2 refactoring ===="
replace_easy1
replace_2

add_try_import
goimports_to_changed

check_build_and_pick

echo "====== complex refactoring ===="

multiline_11
check_build_and_pick

multiline_3
check_build_and_pick

multiline_2
check_build_and_pick

multiline_1
check_build_and_pick

# checking goimports at the end
goimports_to_changed
