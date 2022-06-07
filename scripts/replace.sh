#!/bin/bash

use_perl=${use_perl:-""}
osname="$(~/aws/os-name)"

process_args() {
	if [ "$2" = "" ]; then
		echo "Usage: "$(basename "$0")" <search> <replace>"
		exit 1
	fi
	first="$1"
	if [ "$3" != "" ]; then
		second="$2"
		third="$3"
	else
		second="$1"
		third="$2"
	fi
	sr="s/$second/$third/g"

	if [ ! -z $debug ]; then
		echo MODE:$use_perl $first $sr
	fi
}

do_work() {
	if [[ "$osname" == "Mac" ]] ; then  
		if [ -z $use_perl ]; then
			ag -l "$first" | xargs sed -Ei '' "$sr"
		else 
			ag -l "$first" | xargs perl -i -p0e "$sr"
		fi
	else # Linux, etc.
		if [ -z $use_perl ]; then
			ag -l "$first" | xargs -r sed -Ei "$sr"
		else 
			ag -l "$first" | xargs -r perl -i -p0e "$sr"
		fi
	fi
}

# Execute main() if this is run in standalone mode (i.e. not in a unit test).
ARGV0="$(basename "$0")"
argv0="$(echo "${ARGV0}" |sed 's/_test$//;s/_test\.sh$//')"

if [ "${ARGV0}" = "${argv0}" ]; then
	process_args "$@"
	do_work
fi

