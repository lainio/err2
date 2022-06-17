#!/bin/bash

filtered_build() {
	local osname=$(uname -s)
	local pkg=${1:-"./..."}
	local awk_file="$location"/delete-"$osname".awk

	res=$(go build -o /dev/null "$pkg" 2>&1 >/dev/null) 
	del=$(echo "$res" | grep 'declared but not used' | awk -F : -f "$awk_file") 

	if [[ $del != "" ]]; then
		eval $del
		vlog "filtered"
		echo "FILTER"
	else
		if [[ $res != "" ]]; then
			vlog "BUILD ERR"
			echo "ERR"
		else
			vlog "BUILD OK"
			echo "OK"
		fi
	fi
}


print_env() {
	if [[ "" != $verbose ]]; then
		echo "---------- env setup -----------------"
		echo "start_branch: '$start_branch'"
		echo "migration_branch: '$migration_branch'"
		echo "use_current_branch: '$use_current_branch'"
		echo "only_simple: '$only_simple'"
		echo "---------- env setup -----------------"
	fi
}

vlog() {
	if [[ "" != $verbose ]]; then
		echo "$1"
	fi
}

check_prerequisites() {
	for c in ag perl sed awk git go jq xargs goimports; do
		if ! [[ -x "$(command -v ${c})" ]]; then
			echo "ERR: missing command: '${c}'." >&2
			echo "Please install before continue." >&2
			exit 1
		fi
	done

	if [[ $allow_subdir == "" && $(git rev-parse --show-toplevel 2>/dev/null) != "$PWD" ]]; then
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

	if go build -o /dev/null "$pkg"; then
		echo "OK"
	else
		local result=$(filtered_build "$pkg")
		vlog "$result"
		echo "$result"
	fi
}

check_commit() {
	local bads=""
	for file in $(ag -l "$1" ); do
		vlog "processing: $file"
		perl -i -p0e "s/$1/$2/mg" $file
		# cleaning: '_ := '
		clean_noname_var_assings_1 $file
		clean_orphan_var_1 $file
		if [[ $(fast_build_check $file) == "OK" ]]; then
			commit_one $file
		else
			# revert changes per bad file that we can test builds
			# this is not perfect because we cannot test 'go builds'
			# per file but package and if package is large one bad
			# eg can ruine it.
			bads+="${file} "
			undo_one $file
		fi
	done
	# still change bad apples that we don't lost all the automatic changes
	for file in $bads; do 
		echo "BAD file: $file, update manually!!!" >&2
		perl -i -p0e "s/$1/$2/mg" $file
		# cleaning: '_ := '
		clean_noname_var_assings_1 $file
		clean_orphan_var_1 $file
	done
}

check_build_and_pick() {
	check_dirty
	local bads=""
	for file in $dirty; do
		if [[ $(fast_build_check $file) == "OK" ]]; then
		# if fast_build_check "$file"; then
			vlog "Build OK with with err2 auto-refactoring: $file"
			git commit -m "err2:$file" $file 1>/dev/null
		else
			echo "TODO: manually check file: $file" >&2
			bads+="${file} "
		fi
	done
	echo $bads
}

clean_orphan_var() {
	vlog "Cleaning: var someVar Type\nsomeVar = try.ToX() -> someVar = try.ToX(), for $1"
	"$location"/replace-perl.sh '(^\s*)var (\w*) .*\n(^\s*)\2 = (try\.To1)' '\1\2 := \4'
	"$location"/replace-perl.sh '(^\s*)var (\w*) .*\n(^\s*)_, \2 = (try\.To2)' '\1_, \2 := \4'
	"$location"/replace-perl.sh '(^\s*)var (\w*) .*\n(^\s*)\2, _ = (try\.To2)' '\1\2, _ := \4'
}

clean_orphan_var_1() {
	local file="$1"
	vlog "Cleaning: var someVar Type\nsomeVar = try.ToX() -> someVar = try.ToX(), for $1"
	perl -i -p0e 's/(^\s*)var (\w*) .*\n(^\s*)\2 = (try\.To1)/\1\2 := \4/mg' $file
	perl -i -p0e 's/(^\s*)var (\w*) .*\n(^\s*)_, \2 = (try\.To2)/\1_, \2 := \4/mg' $file
	perl -i -p0e 's/(^\s*)var (\w*) .*\n(^\s*)\2, _ = (try\.To2)/\1\2, _ := \4/mg' $file
}

clean_noname_var_assings_1() {
	local file="$1"
	vlog "Cleaning: _ := try.ToX() -> try.ToX(), for  $1"
	perl -i -p0e 's/(^\s*)(_ :?= )(try\.To1)/\1\3/mg' $file
	perl -i -p0e 's/(^\s*)(_, _ :?= )(try\.To2)/\1\3/mg' $file
	perl -i -p0e 's/(^\s*)(_, _, _ :?= )(try\.To3)/\1\3/mg' $file
}

clean_noname_var_assings() {
	vlog "Cleaning: _ := try.To(... assignments"
	"$location"/replace.sh '(^\s*)(_ :?= )(try\.To1)' '\1\3'
	"$location"/replace.sh '(^\s*)(_, _ :?= )(try\.To2)' '\1\3'
	"$location"/replace.sh '(^\s*)(_, _, _ :?= )(try\.To3)' '\1\3'
}

clean() {
	clean_orphan_var
	clean_noname_var_assings
}

try_3() {
	vlog "Combine one try.To3() call"
	check_commit '(^\s*[\w\.]*, [\w\.]*, [\w\.]*)(, err)( :?= )(.*?)(\n)(\s*try\.To\(err\))' '\1\3try.To3(\4)'
}

multiline_3() {
	vlog "Combine multiline try.To3() calls"
	check_commit '(^\s*\w*, \w*, \w*)(, err)( :?= )((.|\n)*?)(\n)(\s*try\.To\(err\))' '\1\3try.To3(\4)'
}

try_2() {
	vlog "Combine ONE try.To2() call"
	check_commit '(^\s*[\w\.]*, [\w\.]*)(, err)( :?= )(.*?)(\n)(\s*try\.To\(err\))' '\1\3try.To2(\4)'
}

#search_2_multi="(^\s*\w*, \w*)(, err)( :?= )([\s\S]*?)(\n)(\s*try\.To\(err\))"
search_2_multi="(^\s*[\w\.]*, [\w\.]*)(, err)( :?= )([\s\S]*?)(\n)(\s*try\.To\(err\))"
#search_2_multi="(^\s*\w*, \w*)(, err)( :?= )([.\n]*?)(\n)(\s*try\.To\(err\))"

search_2() {
	set +e # if you want to run many search!!
	vlog "Searching: $search_2_multi"
	ag "$search_2_multi"
}

multiline_2() {
	vlog "Combine multiline try.To2() calls"
	#check_commit '(^\s*\w*, \w*)(, err)( :?= )((.|\n)*?)(\n)(\s*try\.To\(err\))' '\1\3try.To2(\4)'
	check_commit "$search_2_multi" '\1\3try.To2(\4)'
}

search_1() {
	set +e # if you want to run many search!!

	vlog "search test $search_1_multi"
	#ag '(^\s*\w*)(, err)( :?= )(.*?)(\n)(\s*try\.To\(err\))'
	#ag '(^\s*(\w|\.)*)(, err)( :?= )((.|\n)*)(\n)(\s*try\.To\(err\))'
	#ag '(^\s*(\w|\.)*)(, err)( :?= )((.|\n)*?)(\n)(\s*try\.To\(err\))'
	ag "$search_1_multi"
}

try_1() {
	vlog "Combine ONE try.To1() calls: to previous lines"
	check_commit '(^\s*[\w\.]*)(, err)( :?= )(.*?)(\n)(\s*try\.To\(err\))' '\1\3try.To1(\4)'
}

search_1_multi='(^\s*[\w\.]*)(, err)( :?= )([\s\S]*?)(\n)(\s*try\.To\(err\))'

multiline_1() {
	vlog "Combine multiline try.To1() calls: following lines"
	vlog "$search_1_multi"
	#check_commit '(^\s*\w*)(, err)( :?= )((.|\n)*?)(\n)(\s*try\.To\(err\))' '\1\3try.To1(\4)'
	check_commit "$search_1_multi" '\1\3try.To1(\4)'
}

try_0() {
	vlog "Combine one err = XXXXX()\ntry.To()"
	check_commit '(^\s*)(err)( :?= )(.*?)(\n)(\s*try\.To\(err\))' '\1try.To(\4)'
}

search_0_multi='(^\s*)(err)( :?= )((.|\n)*?)(\n)(\s*try\.To\(err\))' 

multiline_0() {
	vlog "Combine multiline err = XXXXX()\ntry.To() calls: following lines"
	check_commit "$search_0_multi" '\1try.To(\4)'
}

search_0() {
	vlog "Search-0: $search_0_multi"
	check_commit "$search_0_multi" '\1try.To(\4)'
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

todo2l() {
	vlog "Searching lone: try.To(err) and listing files"
	ag -l  '^\s*try\.To\(err\)$'
}
