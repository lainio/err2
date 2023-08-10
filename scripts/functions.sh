#!/bin/bash

filtered_build() {
	local osname=$(uname -s)
	local pkg=${1:-"./..."}
	local awk_file="$location"/delete-"$osname".awk

	res=$(go build -o /dev/null "$pkg" 2>&1 >/dev/null) 
	del=$(echo "$res" | grep 'declared but not used' | awk -F : -f "$awk_file") 

	if [[ $del != "" ]]; then
		eval $del
		dlog "filter processing working"
		echo "FILTER"
	else
		if [[ $res != "" ]]; then
			echo "ERR"
		else
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

dlog() {
	if [[ "" != $debug ]]; then
		echo "$1" >&2
	fi
}

vlog() {
	if [[ "" != $verbose ]]; then
		echo "$1"
	fi
}

check_prerequisites() {
	for c in ag perl sed awk git go tr jq xargs goimports; do
		if ! [[ -x "$(command -v ${c})" ]]; then
			echo "ERR: missing command: '${c}'." >&2
			echo "Please install before continue." >&2
			exit 1
		fi
	done

	if [[ $allow_subdir == "" && $(git rev-parse --show-toplevel 2>/dev/null) != "$PWD" ]]; then
		echo "ERR: your current dir must be repo's root dir" >&2
		exit 1
	fi

	local go_version=$(go mod edit -json | jq -r '."Go"')
	if [[ $go_version < 1.19 ]]; then
		echo "ERROR:  Go version number ($go_version) is too low" >&2
		echo "Sample: go mod edit -go=1.19 # sets the minimal version" >&2
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

build_all() {
	local pkg=${1:-"./..."}
	go build -o /dev/null "$pkg" 
}

check_build() {
	local pkg="$1"
	go build -o /dev/null "$pkg" &>/dev/null
}

add_assert_import() {
	vlog "add push/pop"
	"$location"/replace-perl.sh '\"github.com\/stretchr\/testify\/(require|assert)\"' '(^func Test)(?!Main)(.*$)' '\1\2\n\tassert.PushTester(t)\n\tdefer assert.PopTester()' 
	vlog "add push/pop for t.Run"
	"$location"/replace.sh '\"github.com\/stretchr\/testify\/(require|assert)\"' '(^\s*)(t\.Run\(.*$)' '\1\2\n\1assert.PushTester(t)\n\1defer assert.PopTester()' 

	vlog "Adding assert|require imports"
	"$location"/replace.sh '\"github.com\/stretchr\/testify\/(require|assert)\"' '\"github.com\/stretchr\/testify\/(require|assert)\"' '\"github.com\/lainio\/err2\/assert\"' 
}

replace_assert() {
	"$location"/replace-perl.sh '(assert|require)\.(Len\()(t,)(.*\))' 'assert.SLen(\4'
	"$location"/replace-perl.sh '(assert|require)\.(NotNil\()(t,)(.*\))' 'assert.INotNil(\4'
	"$location"/replace-perl.sh '(assert|require)\.(Nil\()(t,)(.*\))' 'assert.SNil(\4'
	"$location"/replace-perl.sh '(assert|require)\.(False\()(t,)(.*\))' 'assert.ThatNot(\4'
	"$location"/replace-perl.sh '(assert|require)\.(True\()(t,)(.*\))' 'assert.That(\4'
	"$location"/replace-perl.sh '(assert|require)\.(Equal\()(t,)(.*\))' 'assert.DeepEqual(\4'
	"$location"/replace-perl.sh '(assert|require)\.(\w*\()(t,)(.*\))' 'assert.\2\4'
}

replace_annotate2() {
	"$location"/replace-perl.sh 'defer err2\.Annotate\(' 'defer err2.Handle('

	"$location"/replace-perl.sh 'defer err2\.Annotatew\(' 'defer err2.Handle('
}

replace_return_blain() {
	"$location"/replace-perl.sh 'defer err2\.Return\(\&err\)' 'defer err2.Handle(&err, nil)'
	"$location"/replace-perl.sh 'defer err2\.Returnf\(' 'defer err2.Handle('
	"$location"/replace-perl.sh 'defer err2\.Returnw\(' 'defer err2.Handle('
}

replace_catch() {
	"$location"/replace-perl.sh 'defer err2\.CatchTrace\(' 'defer err2.Catch('
	"$location"/replace-perl.sh 'defer err2\.CatchAll\(' 'defer err2.Catch('
}

replace_return() {
	"$location"/replace-perl.sh 'defer err2\.Return\(' 'defer err2.Handle('
	"$location"/replace-perl.sh 'defer err2\.Returnf\(' 'defer err2.Handle('
	"$location"/replace-perl.sh 'defer err2\.Returnw\(' 'defer err2.Handle('
}

replace_tracers() {
	"$location"/replace-perl.sh 'err2\.(StackTraceWriter = )(.*)' 'err2.SetTracers(\2)'
}

replace_defasserter() {
	"$location"/replace-perl.sh 'assert\.(DefaultAsserter = )(.*)' 'assert.SetDefaultAsserter(\2)'
}

replace_defasserter_prod2() {
	vlog '--- exectuing asserter prod change'
	"$location"/replace-perl.sh 'assert\.SetDefaultAsserter.*\)' 'assert.SetDefault(assert.Production)'
}

replace_defasserter_test2() {
	vlog '--- exectuing asserter Test change'
	"$location"/replace-perl.sh 'func.*testing\.' 'assert\.SetDefaultAsserter.*\)' 'assert.SetDefault(assert.TestFull)'
}

replace_asserters_calls() {
	vlog '--- exectuing asserter.D/P calls'
	"$location"/replace-perl.sh 'assert\.([DP]+\.)' 'assert.' 
	"$location"/replace-perl.sh 'assert\.True' 'assert.That' 
	"$location"/replace-perl.sh 'assert\.Truef' 'assert.That' 
	"$location"/replace-perl.sh 'assert\.NoImplementation' 'assert.NotImplemented' 
	"$location"/replace-perl.sh 'assert\.[DP]+' '(^\s*)(assert\.[DP]+\ \=.*$)' 'EMPTY_THIS_LINE' 
}

replace_annotate() {
	# Replace Annotate with Returnf: notice argument order!!
	"$location"/replace-perl.sh 'err2\.(Annotate\()(.*)(, )(.*)(\)\n)' 'err2.Returnf(\4\3\2\5'

	# Replace Annotatew with Returnf: notice argument order!!
	"$location"/replace-perl.sh 'err2\.(Annotatew\()(.*)(, )(.*)(\)\n)' 'err2.Returnw(\4\3\2\5'
}

replace_err_values() {
	"$location"/replace.sh 'err2\.NotFound' 'err2.ErrNotFound'
	"$location"/replace.sh 'err2\.NotExist' 'err2.ErrNotExist'
	"$location"/replace.sh 'err2\.AlreadyExist' 'err2.ErrAlreadyExist'
	"$location"/replace.sh 'err2\.NotAccess' 'err2.ErrNotAccess'
	"$location"/replace.sh 'err2\.NotRecoverable' 'err2.ErrNotRecoverable'
	"$location"/replace.sh 'err2\.Recoverable' 'err2.ErrRecoverable'
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

	if check_build "$pkg"; then
		echo "OK"
	else
		local result=$(filtered_build "$pkg")
		dlog "$result"
		echo "$result"
	fi
}

check_commit() {
	local bads=""
	for file in $(ag -l "$1" ); do
		echo "processing: $file"
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
			# e.g can ruin it.
			bads+="${file} "
			undo_one $file
		fi
	done
	# still change bad apples that we don't lost all the automatic changes
	for file in $bads; do 
		echo "Problematic file: $file, keeping changes but no commit" >&2
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
			vlog "Build OK with with err2 auto-refactoring: $file"
			git commit -m "err2:$file" $file 1>/dev/null
		else
			echo "The file: $file needs more processing..." >&2
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
	dlog "Cleaning: var someVar Type\nsomeVar = try.ToX() -> someVar = try.ToX(), for $1"
	perl -i -p0e 's/(^\s*)var (\w*) .*\n(^\s*)\2 = (try\.To1)/\1\2 := \4/mg' $file
	perl -i -p0e 's/(^\s*)var (\w*) .*\n(^\s*)_, \2 = (try\.To2)/\1_, \2 := \4/mg' $file
	perl -i -p0e 's/(^\s*)var (\w*) .*\n(^\s*)\2, _ = (try\.To2)/\1\2, _ := \4/mg' $file
}

clean_noname_var_assings_1() {
	local file="$1"
	dlog "Cleaning: _ := try.ToX() -> try.ToX(), for  $1"
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

search_0_multi='(^\s*)(err)( :?= )((.|\n)*?)(\n)(\s*try\.To\(err\))' 
search_1_multi='(^\s*[\w\.]*)(, err)( :?= )([\s\S]*?)(\n)(\s*try\.To\(err\))'
search_2_multi='(^\s*[\w\.]*, [\w\.]*)(, err)( :?= )([\s\S]*?)(\n)(\s*try\.To\(err\))'
search_3_multi='(^\s*[\w\.]*, [\w\.]*, [\w\.]*)(, err)( :?= )([\s\S]*?)(\n)(\s*try\.To\(err\))'

# other tested versions, left here for debugging purposes
#search_2_multi="(^\s*\w*, \w*)(, err)( :?= )([\s\S]*?)(\n)(\s*try\.To\(err\))"
#search_2_multi="(^\s*\w*, \w*)(, err)( :?= )([.\n]*?)(\n)(\s*try\.To\(err\))"

search_3() {
	set +e # if you want to run many search!!
	dlog "Searching: $search_3_multi"
	ag "$search_3_multi"
}

search_2() {
	set +e # if you want to run many search!!
	dlog "Searching: $search_2_multi"
	ag "$search_2_multi"
}

search_1() {
	set +e # if you want to run many search!!

	dlog "search test $search_1_multi"
	ag "$search_1_multi"
}

search_0() {
	dlog "Search-0: $search_0_multi"
	check_commit "$search_0_multi" '\1try.To(\4)'
}

try_3() {
	dlog "Combine ONE try.To3() call"
	check_commit '(^\s*[\w\.]*, [\w\.]*, [\w\.]*)(, err)( :?= )(.*?)(\n)(\s*try\.To\(err\))' '\1\3try.To3(\4)'
}

try_2() {
	dlog "Combine ONE try.To2() call"
	check_commit '(^\s*[\w\.]*, [\w\.]*)(, err)( :?= )(.*?)(\n)(\s*try\.To\(err\))' '\1\3try.To2(\4)'
}

try_1() {
	dlog "Combine ONE try.To1() calls: to previous lines"
	check_commit '(^\s*[\w\.]*)(, err)( :?= )(.*?)(\n)(\s*try\.To\(err\))' '\1\3try.To1(\4)'
}

try_0() {
	dlog "Combine ONE err = XXXXX()\ntry.To()"
	check_commit '(^\s*)(err)( :?= )(.*?)(\n)(\s*try\.To\(err\))' '\1try.To(\4)'
}

multiline_3() {
	dlog "Combine multi-line try.To3() calls: $search_3_multi"
	check_commit "$search_3_multi" '\1\3try.To3(\4)'
}

multiline_2() {
	dlog "Combine multi-line try.To2() calls"
	#check_commit '(^\s*\w*, \w*)(, err)( :?= )((.|\n)*?)(\n)(\s*try\.To\(err\))' '\1\3try.To2(\4)'
	check_commit "$search_2_multi" '\1\3try.To2(\4)'
}


multiline_1() {
	dlog "Combine multi-line try.To1() calls: following lines"
	dlog "$search_1_multi"
	check_commit "$search_1_multi" '\1\3try.To1(\4)'
}

multiline_0() {
	dlog "Combine multi-line err = XXXXX()\ntry.To() calls: following lines"
	check_commit "$search_0_multi" '\1try.To(\4)'
}

check_if_stop_for_simplex() {
	if [[ ! -z $only_simple ]]; then
		exit -1
	fi
}

todo() {
	dlog "Searching err2 references out of catchers"
	ag -l 'err2\.(Check|Try|Filter|CatchAll|CatchTrace|Annotate|Return)'
}

todo_show() {
	dlog "Searching err2 references out of catchers"
	ag 'err2\.(Check|Try|Filter|CatchAll|CatchTrace|Annotate|Return)'
}

todo2() {
	dlog "Searching lone: try.To(err)"
	ag -B 15  '^\s*try\.To\(err\)$'
}

todo2l() {
	dlog "Searching lone: try.To(err) and listing files"
	ag -l  '^\s*try\.To\(err\)$'
}

todo_assert() {
	# TODO: idea, study how to send arguments to these functions when
	# calling them outside of the scripts
	dlog "Searching lone: assert.D/P, no automatic replace yet.."
	ag 'assert\.[DP]+\.'
}

search_handle_multi='(^\s*)(defer err2\.Handle\(&err, func\(\) \{)([\s\S]*?)(^\s*\}\)$)'

todo_handle_func() {
	vlog "searching old error Handlers"
	ag "$search_handle_multi"
}

repl_handle_func() {
	vlog "replacing old error Handlers"
	check_commit "$search_handle_multi" '\1defer err2.Handle(&err, func(err error) error {\3\1\treturn err\n\1})'
}

search_catch_multi='(^\s*)(defer err2\.Catch\(func\(err error\) \{)([\s\S]*?)(^\s*\}\)$)'

todo_catch_func() {
	vlog "searching old error Catchers"
	ag "$search_catch_multi"
}

repl_catch_func() {
	vlog "replacing old error Catchers"
	check_commit "$search_catch_multi" '\1defer err2.Catch(err2.Err(func(err error) {\3\1}))'
}

lint() {
	dlog "Linter check for missing defers"
	ag '^\s*err2\.(Handle|Catch)'
}

lint_ok_handle() {
	dlog "Linter check for missing defers"
	ag '^\s*defer err2\.(Handle|Catch)'
}
