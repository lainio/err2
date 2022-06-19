## Automatic Migration

The err2 doesn't have type variables (`err2.Int.Try(), err2.Bool.Try()`) since
Go generics. They have been deprecated since version 0.8.0. Now they are removed
from the repo as obsolete. Similarly `err2.Check(err error)` is replaced by
`try.To(err error)`

If your projects and repos are using `err2` version before Go generics, you can
migrate them automatically. Just run the `./migrate.sh` bash script. It will do
it for you, and it's safe. It uses git for undo buffer and stored any successful
migration step.

First, to make it more convenient:

```console
mkdir -p $GOPATH/src/github.com/lainio/
cd $GOPATH/src/github.com/lainio/
git clone https://github.com/lainio/err2
cd err2/scripts
source ./scripts/set-path.sh
```

Second, go to your repo's root directory (NOTE, if you have sub-modules read the
[Sub Modules In The Repo](#sub_modules) first!) and enter the following command:

```console
migrate.sh
```

It will create a new `err2-auto-update` branch from your current branch and
change and commits several files to git. If you want to squash separated commits
after everything is done, you have to do it manually after auto-refactoring.
When each changed file is committed separately, it will offer better
problem-solving capabilities for manually fixing things. NOTE: Don't worry;
scripts will change all files successfully, and no manual work is needed, or
it's very minimal.

If the `migrate.sh` script cannot change all the files successfully, it leaves
those files in the git staging area, and you must fix the compilation errors by
yourself. After that, you should run:

```console
migr-name.sh todo # if there is still some basics to migrate like err2.Check(err)
migr-name.sh todo2 # if there is still some complex multiline migrations
```

If there still is something to do, check them manually. They usually are false
positive where the error value has been checked by hand in the old code. At last
step still check that everything compiles:

```console
migr-name.sh build_all # checks that your package builds
```

Before you create a PR, you should run linters and tests. Naturally, you can
diff the migration branch to the original to see what's been happening. That is
an excellent point to check if there is opportunity to use `try.Is` to
simplify code.

### Sub Modules In The Repo

If your git repo has multiple Go modules aka sub-modules, it's recommended that
you start with them. Go to the directory where their `go.mod` files is and run:

```console
migrate.sh -s err2-select-your-migration-branch-name
```

Do that to each of the sub-modules with the **same branch name each time**
before returning back to the root directory. In the root directory enter the
command **once again with the same branch name**:

```console
migrate.sh err2-your-previously-selected-branch-name
```

### Migration Options

Both tools `migrate.sh` and `migr-name.sh` take flags. You can see the usage by:

```console
migrate.sh -h
```

For example, you can give the migration branch an argument to the `migrate.sh`
branch. You could add conversions to the `migrate.sh` script if you have
generated your type variables before running it. The file itself has `TODO` tags
for that.

For debugging and starting some specific migration separately, there is a
`migr-name.sh` helper script that's a starter script that helps you to run all
of the functions in the `functions.sh` script.

We currently support only the `bash` version, which is relatively modern. Also
script automatically checks that all of the needed tools are installed like:
`ag`, `perl`, `git`, etc.

Please let us know what you think, and give us feedback at GitHub Discussions.

## Release Readme

Run the `release.sh` script:

```console
./release.sh <VERSION_TO_RELEASE>
```

Note! The version string format is 'v0.8.0'. Don't forget the v at the
beginning. TODO: update the script.
