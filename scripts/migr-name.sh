#!/bin/bash

location=$(dirname "$BASH_SOURCE")
. $location/functions.sh

start_branch=$(git rev-parse --abbrev-ref HEAD)
migration_branch=${migration_branch:-"err2-update"}
no_build_check=${no_build_check:-"1"}
use_current_branch=${use_current_branch:-"1"}

echo "location: $location"
echo "$BASH_SOURCE"

for a in "$@"; do
	$a
done
