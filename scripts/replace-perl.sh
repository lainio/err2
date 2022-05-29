#!/bin/bash

location=$(dirname "$BASH_SOURCE")
name=$(basename "$location")
[[ "$name" = "." ]] && name=$(basename "$PWD")
echo $location
echo $name

use_perl=perl "$location"/replace.sh "$@"
