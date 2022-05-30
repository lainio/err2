#!/bin/bash

location=$(dirname "$BASH_SOURCE")
echo $location

# # Replace FilterTry with our new version
# "$location"/replace.sh 'err2\..*FilterTry\(' 'try.Is('
# 
# # Replace IsEOF with our new version
# "$location"/replace.sh 'err2\.TryEOF\(' 'try.IsEOF('
# 
# # replace StrStr as it own because it returns two values
# "$location"/replace.sh 'err2\.StrStr\.Try\(' 'try.To2('
# 
# # change all the rest type variable usages.
# # todo: if you have your own of them like e2.XxxxType.Try use this as guide
# "$location"/replace.sh 'err2\..*Try\(' 'try.To1('
# 
# # add try import
# "$location"/replace.sh '\"github.com\/lainio\/err2\"' '\"github.com\/lainio\/err2\"\n\t\"github.com\/lainio\/err2\/try\"\n' 
# 
# "$location"/replace.sh '(err2.Check\()(.*)(\))' 'try.To(\2\3'
# 
# 5 ok:
#use_perl=perl "$location"/replace.sh '(, \w*)(, err)( :?= )([\w\(\)\[\],\. ]*)(\n)(\s*try.To\(err\))' '\1\3try.To2(\4)'
#"$location"/replace-perl.sh '(, \w*)(, err)( :?= )([\w\(\)\[\],\. "]*)(\n)(\s*try.To\(err\))' '\1\3try.To2(\4)'

# 6 ok:
#"$location"/replace-perl.sh '(, err)( :?= )([\w\s\.,"\(\)\{\}\[\]\*]*)(\n)(\s*try.To\(err\))' '\2try.To1(\3)'
"$location"/replace-perl.sh '(, err)( :?= )([\w\s\.,:;%&\(\)\{\}\[\]\$\^\?\\\|\+]*)(\n)(\s*try.To\(err\))' '\2try.To1(\3)'

# err :?=[\w\s\.,\(\)\{\}\[\]\*]*try\.To
# cleanup and add needed imports
#goimports -l -w .

# last two are space and " remember ' or not!
# WORKING: \w\(\)\[\],\.\n\t\{\} "
# New:
# err :?=[\w\s\.,:;%&\(\)\{\}\[\]\$\^\?\\\|\+]*try\.To
