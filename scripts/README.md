## Automatic Migration

The err2 package will offer auto-migration scripts until version 1.0.0 is
published. That means that you can safely use the package even the API is not
yet 100% staple. We have used that approach for our production code that uses
`err2` and it works very well. Over 100 KLOC Go code has been auto-migrated
successfully from type variables (see below) to Go generics API.

This readme will guide you to use auto-migration scripts when ever we deprecate
functions or make something obsolete.

### Auto-Migration of v0.9.40

The version 0.9.40 is a major update because of the performance and API change.
We have managed to eliminate `defer` slowdown. Our benchmarks are 3x faster than
previous version and about equal to those function call stacks (100 level) that
don't use `defer`. We are exactly at the same level of performance with those
functions that use deferred function that accept an argument. Transporting an
argument to deferred function seems to be slower than functions that don't. And
because `try.To` is already as fast as `if err != nil` we reached our goal for
speed!

Because all of the error handler function signatures are now:
```go
	ErrorHandler = func(err error) error
```
and the old one was `func()` relaying closures. And because we have dynamic
signatures on `err2.Handle` and `err2.Catch`, we cannot use compiler (Go lacks
the function overloading) for type checking, which makes this migration so
important but difficult. The `err2` package writes `fatal error: ...` to
`stderr` because it cannot panic in middle of the panicking.

> Note. This is only a legacy code problem. We are in the migration as you can
> see :-)

You have two (2) options for migration for v0.9.40:
1. [(Semi)-Automatic Migration](#semi-automatic-migration)
1. [Manual migration using (vim/nvim) location lists or similar](manual-migration-with-a-location-list)

#### (Semi)-Automatic Migration

Follow these steps:
1. [Set up Migration Environment](#set-up-migration-environment)
1. Make sure that you haven't uncommitted changes in your repo and that you are
   in the branch where you want to make the changes.
1. Execute following command in your repo's root directory:
   ```shell
   migr-name.sh -n repl_handle_func repl_catch_func
   ```
1. Then continue with `build`, `test`, `lint`, etc. **And don't stop yet.**
1. Use `git diff` or similar to skimming that all changes seem to be OK. Here
   you have an opportunity to start use new features of `err2` like logging.
1. You are ready to commit changes.

#### Manual Migration With a Location List

Follow these steps check do you have migration needs for v0.9.40:
1. [Set up Migration Environment](#set-up-migration-environment)
1. Execute following command in your repo's root directory:
   ```shell
   migr-name.sh -n todo_handle_func todo_catch_func
   ```

*Tip. If your repo is large and you have many migration changes see the next
[semi-automatic guide](#semi-automatic-migration).*

*Tip. You can execute that command from e.g. nvim/vim and you get your fix list.*
```shell
migr-name.sh todo_handle_func todo_catch_func > todo_list
nvim -q todo_list
```
*Or depending on your current shell:*
```shell
nvim -q <(migr-name.sh todo_handle_func todo_catch_func)
```

### `assert.SetDefaultAsserter` -> `assert.SetDefault` and others in v0.9.1

Because direct renaming causes braking changes remember add -x flag to your
migration command but other follow the instructions below:

Example of -x flag:
```shell
migrate.sh -x
```

Use the following commands because 
### `err2.NotFound`, `err2.NotExist` and other sentinels are renamed in v0.9.0

Their names follow Go idiom even the `err` part is two times here:
`err2.ErrNotFound`.

Please follow the same steps presented for v0.8.10 below to automatically
refactor all references to these error values in your repos. **But add the flag
-x to call**. Example:

```shell
migrate.sh -x
```

### `err2.Return[f/w]` will obsolete in v0.9.0

Please follow the same steps as the next chapter.

### `err2.Annotate` and `err2.StackTraceWriter` are obsolete in v0.8.10

Please follow these guides to automatically replace all obsolete `err2.Annotate`
functions and `err2.StackTraceWriter` variable set with proper API:

1. [Set up migration environment](#set-up-migration-environment)
2. Execute following command to replace all `err2.Annotate(w)` with proper
   `err2.Returnf/w` function:
   `migrate.sh -o` or `migrate.sh -o YOUR-OWN-BRANCH-NAME`

### Type Variables Are Obsolete

The err2 doesn't have type variables (`err2.Int.Try(), err2.Bool.Try()`, etc.)
since Go generics. They have been deprecated as of version 0.8.0. Now they are
removed from the repo as obsolete. Similarly `err2.Check()` is replaced by
`try.To()`

If your projects and repos are using `err2` version before Go generics, you can
migrate them automatically. Just run the `migrate.sh` bash script. It will do it
for you, and it's safe. It uses git for undo buffer. It stores all successful
migration steps to git in its own working branch.

First, to make the use of it more convenient:

#### Set up Migration Environment

1. Clone the `err2` if you don't already have it:
   ```console
   mkdir -p $GOPATH/src/github.com/lainio/
   cd $GOPATH/src/github.com/lainio/
   git clone https://github.com/lainio/err2
   ```

2. Use `set-path.sh` to add scripts directory to path:
   ```console
   cd err2/scripts
   source ./set-path.sh
   ```
   or this works as well:
   ```console
   source <RELATIVE_PATH_TO>/err2/scripts/set-path.sh
   ```

Second, to update your repo for latest `err2` interface.

#### Automatic Version Migration

Go to your repo's root directory (NOTE, if you have sub-modules read the
[Sub Modules In The Repo](#sub-modules-in-the-repo) first!) and enter the
following command:

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
