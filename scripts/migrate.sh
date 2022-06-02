#!/bin/bash

check_prerequisites() {
	for c in ag perl sed git go jq xargs; do
		if ! [ -x "$(command -v ${c})" ]; then
			echo "ERR: missing command: '${c}'."
			echo "Please install before continue." >&2
			exit 1
		fi
	done

	local go_version=$(go mod edit -json | jq -r '."Go"')
	if [ $go_version \< 1.18 ]; then
		echo "ERR: Go version number ($go_version) is too low" >&2
		exit 1
	fi

	if [[ ! -z "$(git status --porcelain)" ]]; then
		echo "ERR: your current branch must be clean" >&2
		exit 1
	fi
}

set -e
location=$(dirname "$BASH_SOURCE")
migration_branch="err2-auto-update"
no_build_check=${no_build_check:-""}

check_dirty() {
	dirty=""
	if [ -z $no_build_check ]; then
		dirty=$(git diff --name-only)
	fi
}

setup_repo() {
	git checkout -b "$migration_branch"
}

deps() {
	go get github.com/lainio/err2
}

check_build() {
	if [ -z $no_build_check ]; then
		go build ./...
	fi
}

replace_easy1() {
	# Replace FilterTry with our new version: notice argument order!!
	"$location"/replace-perl.sh '(err2\.FilterTry\()(.*)(, )(.*)(\)\n)' 'try.Is(\4\3\2\5'

	# Use IsEOF instead of TryEOF
	"$location"/replace.sh 'err2\.TryEOF\(' 'try.IsEOF('

	# replace StrStr as it own because it returns two values
	"$location"/replace.sh 'err2\.StrStr\.Try\(' 'try.To2('
}

replace_1() {
	# change all the rest type variable usages.
	# TODO: if you have your own of them like e2.XxxxType.Try use this as guide
	"$location"/replace.sh 'err2\.\w*\.Try\(' 'try.To1('
}

replace_2() {
	# This is very RARE, remove is you have problems!!!
	"$location"/replace-perl.sh '(err2\.Try\()(\w*?\.)(Read|Fprint|Write)' 'try.To1(\2\3'

	# replace very rare err2.Try() call 
	"$location"/replace.sh '\s*(err2\.Try\()' 'try.To('

	"$location"/replace.sh '(err2.Check\()(.*)(\))' 'try.To(\2\3'
}

add_try_import() {
	"$location"/replace.sh '(try\.To|try\.Is)' '\"github.com\/lainio\/err2\"' '\"github.com\/lainio\/err2\"\n\t\"github.com\/lainio\/err2\/try\"' 
}

goimports_to_changed() {
	git diff --name-only | grep '^.*\.go$' | xargs goimports -l -w
}

commit() {
	check_dirty
	if [[ ! -z "$dirty" ]]; then
		git commit -am "automatic err2 migration: $1"
	fi
}

multiline_3() {
	"$location"/replace-perl.sh '(, \w*)(, \w*)(, err)( :?= )([\w\s\.,:;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\1\2\4try.To3(\5)'
}

multiline_2() {
	"$location"/replace-perl.sh '(, \w*)(, err)( :?= )([\w\s\.,:;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\1\3try.To2(\4)'
}

multiline_1() {
	"$location"/replace-perl.sh '(, err)( :?= )([\w\s\.,:;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'
}

# =================== main =====================

check_prerequisites

echo "update err2 package to latest version"
setup_repo
deps
check_build
commit "commit deps"

echo "====== basic err2 refactoring starts now ===="
echo "calling easy 1"
replace_easy1
echo "calling  1"
replace_1
echo "calling  2"
replace_2
echo "add try imports"
add_try_import
echo "fmt with goimports"
goimports_to_changed
echo "check build"
check_build
echo "commit phase1"
commit "phase 1"

echo "====== complex refactoring starts now ===="
echo "multiline_3"
multiline_3
check_build
commit "phase 2 multilines"

echo "multiline_2"
multiline_2
check_build
commit "phase 2 multilines"

echo "multiline_1"
multiline_1
check_build
commit "phase 2 multilines"

goimports_to_changed
check_build
commit "phase 2 multilines"

