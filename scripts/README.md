## Automatic Migration

The err2 doesn't have type variables any more. They have been deprecated since
version 0.8.0. No they are removed for the repo as obsolete.

They were generated for performance reasons and convenience. You could write:

```go
b := err2.Bytes.Try(ioutil.ReadAll(r))
```

Instead of:

```go
b, err := ioutil.ReadAll(r)
err2.Check(err)
```

Thanks for the Go generics we can write it for any type:

```go
b := try.To1(ioutil.ReadAll(r))
```

You can migrate your `err2` using repos from version which were using a version
of `err2` which didn't have Go generics. You don't need to do it manually but
just run the `./migrate.sh` bash script. It will do it for you.

Go to your repo and enter the following command:

```console
$GOPATH/src/github.com/lainio/err2/scripts/migrate.sh
```

```console
no_build_check=1 $GOPATH/src/github.com/lainio/err2/scripts/migrate.sh
```

```console
use_current_branch $GOPATH/src/github.com/lainio/err2/scripts/migrate.sh
```

### Release Readme

Run the `release.sh` script:

```console
./release.sh <VERSION_TO_RELEASE>
```

Note! The version string format is 'v0.8.0'. Don't forget the v at the
beginning. TODO: update the script.

