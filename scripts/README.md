## Automatic Migration

The err2 doesn't have type variables anymore. They have been deprecated since
version 0.8.0. Now they are removed from the repo as obsolete. Similarly
`err2.Check(err error)` is replaced by `try.To(err error)`

You can migrate your `err2` using repos automatically. Just run the
`./migrate.sh` bash script. It will do it for you.

Go to your repo's root dir and enter the following command:

```console
$GOPATH/src/github.com/lainio/err2/scripts/migrate.sh
```

It will create a new `err2-auto-update` branch from your current branch and
change and commits several files to git. If you want to merge separated commits,
you have to do it manually after auto-refactoring is done. When each changed
file is committed separately, it will offer better problem-solving capabilities
for manually fixing things. NOTE: Don't worry; scripts will change all files
successfully, and no manual work is needed, or it's minimal.

If the `migrate.sh` script cannot change all the files in one run, it leaves
those files in the git staging area, and you must fix the compilation errors by
hand. After that, you should run the `migrate.sh` again until everything
compiles and all files are committed to git.

Before you create a PR, you should run linters and tests. Naturally, you can
diff the migration branch to the original to see what's been done. That is an
excellent point to check if there is a new `err2` API like `try.Is` to simplify
code.

##### Migration Options

You can give the migration branch an argument to the `migrate.sh` branch. You
could add conversions to the `migrate.sh` script if you have generated your type
variables before running it. The file itself has `TODO` tags for that.

For debugging, there is a `migr-name.sh` helper script that's a starter script
that helps you to run all of the functions in the `functions.sh` script.

We currently support only the `bash` version, which is relatively modern. Also
script automatically checks that all of the needed tools are installed like:
`ag`, `perl`, `git`, etc.

Please let us know what you think, and give us feedback at GitHub Discussions.

### Release Readme

Run the `release.sh` script:

```console
./release.sh <VERSION_TO_RELEASE>
```

Note! The version string format is 'v0.8.0'. Don't forget the v at the
beginning. TODO: update the script.
