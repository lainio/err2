#!/bin/bash

location=$(dirname "$BASH_SOURCE")
. $location/functions.sh

migration_branch=${1:-"err2-update"}
no_build_check=${no_check:-"1"}
use_current_branch=${use_current_branch:-"1"}

print_env
"$@"
