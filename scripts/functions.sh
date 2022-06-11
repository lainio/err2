#!/bin/bash

print_env() {
	if [[ "" != $verbose ]]; then
		echo "---------- env setup -----------------"
		echo "start_branch: $start_branch"
		echo "migration_branch: $migration_branch"
		echo "use_current_branch: $use_current_branch"
		echo "only_simple: $only_simple"
		echo "---------- env setup -----------------"
	fi
}

vlog() {
	if [[ "" != $verbose ]]; then
		echo "$1"
	fi
}

check_prerequisites() {
	for c in ag perl sed git go jq xargs goimports; do
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
		echo "ERROR:  Go version number ($go_version) is too low" >&2
		echo "Sample: go mod edit -go=1.18 # sets the minimal version" >&2
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
	vlog "Replacing err2.Check, err2.FilterTry, err2.TryEOF, and type vars"

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
	# TODO: if you have your own of them like e2.XxxxType.Try use this as guide
	#"$location"/replace.sh 'err2\.\w*\.Try\(' 'try.To1('
}

replace_2() {
	vlog "Replacing err2.Try calls"
	# This is very RARE, remove if you have problems!!!
	"$location"/replace-perl.sh '(err2\.Try\()(\w*?\.)(Read|Fprint|Write)' 'try.To1(\2\3'

	# replace very rare err2.Try() call 
	"$location"/replace.sh '\s*(err2\.Try\()' 'try.To('
}

add_try_import() {
	vlog "Adding try imports"
	"$location"/replace.sh '(try\.To|try\.Is)' '\"github.com\/lainio\/err2\"' '\"github.com\/lainio\/err2\"\n\t\"github.com\/lainio\/err2\/try\"' 
}

goimports_to_changed() {
	vlog "Checking with goimports"
	git diff --name-only | grep '^.*\.go$' | xargs goimports -w
}

commit() {
	check_dirty
	if [[ ! -z "$dirty" ]]; then
		git commit -am "automatic err2 migration: $1"
	else
		vlog "All OK, nothing to commit at this phase, continuing checks..."
	fi
}

undo_one() {
	local file="$1"
	if [[ -z "$no_commit" ]]; then
		git checkout -- $file
	fi
}

commit_one() {
	local file="$1"
	if [[ -z "$no_commit" ]]; then
		git commit -m "err2:$file" $file 1>/dev/null
	fi
}

fast_build_check() {
	local pkg="./$(dirname ${1})/..."
	go build -o /dev/null "$pkg"
}

check_commit() {
	local bads=""
	for file in $(ag -l "$1" ); do
		vlog "processing: $file"
		perl -i -p0e "s/$1/$2/mg" $file
		# cleaning: '_ := '
		perl -i -p0e "s/(_ :?= )(try\.To1)/\2/mg" $file
		if fast_build_check $file; then
			commit_one $file
		else
			bads+="${file} "
			undo_one $file
		fi
	done
	for file in $bads; do 
		echo "BAD file: $file, update manually!!!" >&2
		perl -i -p0e "s/$1/$2/mg" $file
		# cleaning: '_ := '
		perl -i -p0e "s/(_ :?= )(try\.To1)/\2/mg" $file
	done
}

check_build_and_pick() {
	check_dirty
	local bads=""
	for file in $dirty; do
		if fast_build_check "$file"; then
			vlog "Build OK with with err2 auto-refactoring: $file"
			git commit -m "err2:$file" $file 1>/dev/null
		else
			echo "TODO: manually check file: $file" >&2
			bads+="${file} "
		fi
	done
	echo $bads
}

clean() {
	vlog "Cleaning: _ := try.To(... assignments"
	"$location"/replace.sh '(^\s*)(_ :?= )(try\.To1)' '\1\3'
	"$location"/replace.sh '(^\s*)(_, _ :?= )(try\.To2)' '\1\3'
	"$location"/replace.sh '(^\s*)(_, _, _ :?= )(try\.To3)' '\1\3'
}

multiline_3() {
	vlog "Combine multiline try.To3() calls"
	check_commit '(^\s*\w*, \w*, \w*)(, err)( :?= )((.|\n)*?)\)(\n)(\s*try\.To\(err\))' '\1\3try.To3(\4))'
	clean
}

multiline_2() {
	vlog "Combine multiline try.To2() calls"
	check_commit '(^\s*\w*, \w*)(, err)( :?= )((.|\n)*?)\)(\n)(\s*try\.To\(err\))' '\1\3try.To2(\4))'
}

# TODO: not used
# testing version
multiline_0() {
	vlog "Combine multiline try.To1() calls: following lines"
	# make a version whichi first change those who has two lines at a row!!
	check_commit '(, |)(err)( :?= )([\w\ \.,:!;%&=\/\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\3try.To1(\4)'

	vlog "Combine multiline try.To1() calls: unlimeted lines"
	check_commit '(, |)(err)( :?= )([\w\s\.,:!;%&=\/\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\3try.To1(\4)'
}

multiline_1() {
	vlog "Combine multiline try.To1() calls: following lines"
	check_commit '(^\s*\w*)(, err)( :?= )(.*?)\)(\n)(\s*try\.To\(err\))' '\1\3try.To1(\4))'
	check_commit '(^\s*\w*)(, err)( :?= )((.|\n)*?)\)(\n)(\s*try\.To\(err\))' '\1\3try.To1(\4))'
}

multiline_11() {
	vlog "Combine multiline err = XXXXX()\ntry.To() calls: following lines"
	# make a version whichi first change those who has two lines at a row!!
	#check_commit '(^\s*err)( :?= )([\w\ \.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To(\3)'
	check_commit '(^\s*)(err :?= )(.*)(\n)(\s*try\.To\(err\))' '\1try.To(\3)'

	#vlog "Combine multiline try.To1() calls: unlimeted lines"
	#check_commit '(, err)( :?= )([\w\s\.,:!;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'
}

check_if_stop_for_simplex() {
	if [[ ! -z $only_simple ]]; then
		exit -1
	fi
}

todo() {
	vlog "Searching err2 references out of catchers"
	ag -l 'err2\.(Check|Try|Filter)'
}

todo_show() {
	vlog "Searching err2 references out of catchers"
	ag 'err2\.(Check|Try|Filter)'
}

todo2() {
	vlog "Searching lone: try.To(err)"
	ag -B 15  '^\s*try\.To\(err\)$'
}
