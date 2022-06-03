## Automatic Migration

You can migrate your `err2` using repos from version which is using a version
which didn't had a Go generics based API. You don't need to do it manually but
just run the `./migrate.sh` bash script. It will do it for you.

Go to your repo and enter the following command:

```console
$GOPATH/src/github.com/lainio/err2/scripts/migrate.sh
```

```console
no_build_check=1 $GOPATH/src/github.com/lainio/err2/scripts/migrate.sh
```

```console
use_current_branch no_build_check=1 $GOPATH/src/github.com/lainio/err2/scripts/migrate.sh
```

### Release Readme

Run the `release.sh` script:

```console
./release.sh <VERSION_TO_RELEASE>
```

Note! The version string format is 'v0.8.0'. Don't forget the v at the
beginning. TODO: update the script.

