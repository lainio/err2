#!/bin/bash

location=$(dirname "$BASH_SOURCE")

# debug=1 use_perl=perl "$location"/replace.sh "$@"
use_perl=perl "$location"/replace.sh "$@"
