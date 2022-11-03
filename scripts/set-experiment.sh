#!/bin/bash

if [[ $(go env GOVERSION) < "go1.20.0" ]]; then
	echo setting go epriment flag
	export GOEXPERIMENT=unified
fi
