#!/bin/bash

print_env() {
	echo "---------- env setup -----------------"
	echo "start_branch: $start_branch"
	echo "migration_branch: $migration_branch"
	echo "use_current_branch: $use_current_branch"
	echo "---------- env setup -----------------"
}

check_prerequisites() {
	for c in ag perl sed git go jq xargs; do
		if ! [[ -x "$(command -v ${c})" ]]; then
			echo "ERR: missing command: '${c}'." >&2
			echo "Please install before continue." >&2
			exit 1
		fi
	done

	if [[ $(git rev-parse --show-toplevel 2>/dev/null) != "$PWD" ]]; then
		echo "ERR: your current dir must be repo's rood dir" >&2
		exit 1
	fi

	local go_version=$(go mod edit -json | jq -r '."Go"')
	if [[ $go_version < 1.18 ]]; then
		echo "ERR: Go version number ($go_version) is too low" >&2
		exit 1
	fi

	if [[ $migration_branch != $start_branch && ! -z "$(git status --porcelain)" ]]; then
		echo "ERR: your current branch must be clean or = '$migration_branch'" >&2
		exit 1
	fi
}

check_dirty() {
	dirty=$(git diff --name-only)
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
	go build -o /dev/null ./...
}

replace_easy1() {
	echo "calling easy 1"

	"$location"/replace.sh 'err2\.Check\(' 'try.To('

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
	#"$location"/replace.sh 'err2\.\w*\.Try\(' 'try.To1('
}

replace_2() {
	echo "calling  2"
	# This is very RARE, remove if you have problems!!!
	"$location"/replace-perl.sh '(err2\.Try\()(\w*?\.)(Read|Fprint|Write)' 'try.To1(\2\3'

	# replace very rare err2.Try() call 
	"$location"/replace.sh '\s*(err2\.Try\()' 'try.To('
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

fast_build_check() {
	local pkg="./$(dirname ${1})/..."
	echo "fast build check: $pkg"
	go build -o /dev/null "$pkg"
}

check_commit() {
	local bads=""
	for file in $(ag -l "$1" ); do
		perl -i -p0e "s/$1/$2/g" $file
		# cleaning: '_ := '
		perl -i -p0e "s/(_ :?= )(try\.To1)/\2/g" $file
		if fast_build_check $file; then
			git commit -m "err2:$file" $file
		else
			bads+="${file} "
			git checkout -- $file
		fi
	done
	for file in $bads; do 
		echo "BAD file: $file, update manually!!!"
		perl -i -p0e "s/$1/$2/g" $file
		# cleaning: '_ := '
		perl -i -p0e "s/(_ :?= )(try\.To1)/\2/g" $file
	done
}

check_build_and_pick() {
	check_dirty
	local bads=""
	for file in $dirty; do
		if fast_build_check "$file"; then
			echo "build ok with update: $file"
			git commit -m "err2:$file" $file
		else
			echo "TODO: manually check file: $file"
			bads+="${file} "
		fi
	done
	echo $bads
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
	#go build ./...

	echo "multiline_1"
	# make a version whichi first change those who has two lines at a row!!
	check_commit '(, err)( :?= )([\w\ \.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'

	check_commit '(, err)( :?= )([\w\s\.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'
}

todo() {
	echo "searching err2 references out of catchers"
	ag -l 'err2\.(Check|Try|Filter)'
}
