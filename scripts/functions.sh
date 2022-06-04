#!/bin/bash

print_env() {
	echo "---------- env setup -----------------"
	echo "start_branch: $start_branch"
	echo "migration_branch: $migration_branch"
	echo "no_build_check: $no_build_check"
	echo "use_current_branch: $use_current_branch"
	echo "---------- env setup -----------------"
}

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

	if [[ $migration_branch != $start_branch && ! -z "$(git status --porcelain)" ]]; then
		echo "ERR: your current branch must be clean or = '$migration_branch'" >&2
		exit 1
	fi
}

check_dirty() {
	dirty=""
	if [[ -z $no_build_check && $migration_branch != $start_branch ]]; then
		dirty=$(git diff --name-only)
	fi
}

setup_repo() {
	if [[ $migration_branch != $start_branch ]]; then
		git checkout -b "$migration_branch"
	fi
}

deps() {
	go get github.com/lainio/err2
}

check_build() {
	if [ -z $no_build_check ]; then
		go build -o /dev/null ./...
		echo "build check OK" 
	else
		echo "skipping build check" 
	fi
}

replace_easy1() {
	echo "calling easy 1"
	# Replace FilterTry with our new version: notice argument order!!
	"$location"/replace-perl.sh '(err2\.FilterTry\()(.*)(, )(.*)(\)\n)' 'try.Is(\4\3\2\5'

	# Use IsEOF instead of TryEOF
	"$location"/replace.sh 'err2\.TryEOF\(' 'try.IsEOF('

	# replace Type Variable helpers as it own because it returns two values
	"$location"/replace.sh 'err2\.StrStr\.Try\(' 'try.To2('
	"$location"/replace.sh 'err2\.R\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.W\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Bools\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Bool\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.File\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Ints\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Strings\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.URL\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Int\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Bytes\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.String\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Int\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Byte\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Empty\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Request\.Try\(' 'try.To1('
	"$location"/replace.sh 'err2\.Response\.Try\(' 'try.To1('

	# TODO: add your own generated type helpers here!!
}

replace_1() {
	echo "calling  1"
	# change all the rest type variable usages.
	# TODO: if you have your own of them like e2.XxxxType.Try use this as guide
	"$location"/replace.sh 'err2\.\w*\.Try\(' 'try.To1('
}

replace_2() {
	echo "calling  2"
	# This is very RARE, remove is you have problems!!!
	"$location"/replace-perl.sh '(err2\.Try\()(\w*?\.)(Read|Fprint|Write)' 'try.To1(\2\3'

	# replace very rare err2.Try() call 
	"$location"/replace.sh '\s*(err2\.Try\()' 'try.To('

	"$location"/replace.sh '(err2.Check\()(.*)(\))' 'try.To(\2\3'
}

add_try_import() {
	echo "add try imports"
	"$location"/replace.sh '(try\.To|try\.Is)' '\"github.com\/lainio\/err2\"' '\"github.com\/lainio\/err2\"\n\t\"github.com\/lainio\/err2\/try\"' 
}

goimports_to_changed() {
	echo "fmt with goimports"
	git diff --name-only | grep '^.*\.go$' | xargs goimports -l -w
}

commit() {
	check_dirty
	if [[ ! -z "$dirty" ]]; then
		git commit -am "automatic err2 migration: $1"
	else
		echo "skipping commit"
	fi
}

check_commit() {
	echo "++ start check commit"
	local goods=""
	local bads=""
	for file in $(ag -l "$1" ); do
		echo "--> perl for $file"
		perl -i -p0e "s/$1/$2/g" $file
		if go build -o /dev/null ./... ; then
			echo "build ok with updated: $file"
			#goods+="${file}\n"
			#git commit -m "err2:$file" $file
		else
			echo "TODO: manually check file: $file"
			bads+="${file} "
			git checkout -- $file
		fi
		echo "next file"
	done
	if go build -o /dev/null ./... ; then
		git commit -am "err2 generator group commit"
	else
		echo "TODO: manually check file: $file"
	fi
	for file in $bads; do 
		echo ">> really update BAD file: $file"
		perl -i -p0e "s/$1/$2/g" $file
	done
}

clean() {
	echo "running clean:"
	"$location"/replace.sh '(_ :?= )(try\.To1)' '\2'
}

multiline_3() {
	echo "multiline_3"
	"$location"/replace-perl.sh '(, \w*)(, \w*)(, err)( :?= )([\w\s\.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\1\2\4try.To3(\5)'
	clean
}

multiline_2() {
	echo "multiline_2"
	"$location"/replace-perl.sh '(, \w*)(, err)( :?= )([\w\s\.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\1\3try.To2(\4)'
	clean
}

multiline_1() {
	go build ./...

	echo "multiline_1"
	# make a version whichi first change those who has two lines at a row!!
	#"$location"/replace-perl.sh '(, err)( :?= )([\w\ \.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'
	#check_commit '(, err)( :?= )([\w\ \.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'
	#clean
	#check_build
	#commit "multiline 1: two lines"

	check_commit '(, err)( :?= )([\w\s\.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'
	#"$location"/replace-perl.sh '(, err)( :?= )([\w\s\.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'
	clean
}

todo() {
	# Catch Return Handle Annotate StackTraceWriter
	ag 'err2\.[^CRHAS]'
}
