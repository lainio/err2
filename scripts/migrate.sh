#!/bin/bash

if [[ ! -z "$(git status --porcelain)" ]]; then
	echo "ERR: your current branch must be clean" >&2
	exit 1
fi

for c in ag perl sed git xargs; do
	if ! [ -x "$(command -v ${c})" ]; then
		echo "ERR: missing command: '${c}'."
		echo "Please install before continue." >&2
		exit 1
	fi
done

set -e

git checkout -b err2-migration

location=$(dirname "$BASH_SOURCE")
# echo $location

# Replace FilterTry with our new version
"$location"/replace.sh 'err2\..*FilterTry\(' 'try.Is('

# Use IsEOF instead of TryEOF
"$location"/replace.sh 'err2\.TryEOF\(' 'try.IsEOF('

# replace StrStr as it own because it returns two values
"$location"/replace.sh 'err2\.StrStr\.Try\(' 'try.To2('

# change all the rest type variable usages.
# todo: if you have your own of them like e2.XxxxType.Try use this as guide
"$location"/replace.sh 'err2\..*Try\(' 'try.To1('

"$location"/replace.sh '(err2.Check\()(.*)(\))' 'try.To(\2\3'

# add try import
"$location"/replace.sh '(try\.To|try\.Is)' '\"github.com\/lainio\/err2\"' '\"github.com\/lainio\/err2\"\n\t\"github.com\/lainio\/err2\/try\"' 

# ============
# == ff here =
git diff --name-only | xargs goimports -l -w
# ============

go build -o /dev/null ./...
git commit -am 'automatic err2 migration phase 1 ok'

# === three return values, don't use yet
#"$location"/replace-perl.sh '(, \w*)(, \w*)(, err)( :?= )([\w\s\.,:;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\1\2\4try.To3(\5)'

#use_perl=perl "$location"/replace.sh '(, \w*)(, err)( :?= )([\w\(\)\[\],\. ]*)(\n)(\s*try.To\(err\))' '\1\3try.To2(\4)'
#"$location"/replace-perl.sh '(, \w*)(, err)( :?= )([\w\(\)\[\],\. "]*)(\n)(\s*try.To\(err\))' '\1\3try.To2(\4)'
# NEW, latest version ====== this the SECOND
"$location"/replace-perl.sh '(, \w*)(, err)( :?= )([\w\s\.,:;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\1\3try.To2(\4)'

# '(, \w*)(, err)( :?= )([\w\s\.,:;%&\-\(\){}\[\]\$\^\?\\\|\+\"\*]*)(\s*try\.To\(err\))'
# 6 ok:
#"$location"/replace-perl.sh '(, err)( :?= )([\w\s\.,"\(\)\{\}\[\]\*]*)(\n)(\s*try.To\(err\))' '\2try.To1(\3)'
#"$location"/replace-perl.sh '(, err)( :?= )([\w\s\.,:;%&-\*\(\)\{\}\[\]\$\^\?\\\|\+\"]*)(\n)(\s*try.To\(err\))' '\2try.To1(\3)'
"$location"/replace-perl.sh '(, err)( :?= )([\w\s\.,:;%&=\-\(\)\{\}\[\]\$\^\?\\\|\+\"\*]*?)(\n)(\s*try\.To\(err\))' '\2try.To1(\3)'

# err :?=[\w\s\.,\(\)\{\}\[\]\*]*try\.To
# cleanup and add needed imports
#goimports -l -w .

git diff --name-only | xargs goimports -l -w

go build -o /dev/null ./...

git commit -am 'automatic err2 migration phase 2 ok'

