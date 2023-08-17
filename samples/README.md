# Samples

Please play with the samples by editing them and running the main files:
- `main.go` just a starter for different playgrounds
- `main-play.go` general playground based on CopyFile and recursion
- `main-db-sample.go` simulates DB transaction and money transfer
- `main-nil.go` samples and tests for logger and using `err2.Handle` for success

Run a default playground `play` mode:
```go
go run ./...
```

Or run the DB based version to maybe better understand how powerful the
automatic error string building is:
```go
go run ./... -mode db
```

You can print usage:
```go
go run ./... -h
```
