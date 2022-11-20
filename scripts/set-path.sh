#!/bin/bash

location=$(dirname "$BASH_SOURCE")
echo $PWD/$location
export PATH=$PATH:$PWD/$location

