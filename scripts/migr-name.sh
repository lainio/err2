#!/bin/bash

location=$(dirname "$BASH_SOURCE")

. $location/migrate.sh

clean
#"$@"
