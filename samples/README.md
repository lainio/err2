# Samples

Please play with the samples by editing them and running the main files:
- `main.go` just a starter for different playgrounds (includes `asserter` tester)
- `main-play.go` general playground based on `CopyFile` and recursion
- `main-db-sample.go` simulates DB transaction with money transfer
- `main-nil.go` samples and tests for logger and using `err2.Handle` for success

Run a default playground `play` mode:
```go
go run .
```

> [!TIP]
> Set a proper alias to play with samples:
> ```sh
> alias sa='go run .'
> ```

Run the DB based version to maybe better understand how powerful the automatic
error string building is:

```go
sa -mode db
# or
go run . -mode db
```

You can print usage:
```go
sa -h
# or
go run . -h
```
