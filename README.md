# Greeter

Test using gripmock SDK (embedded mock).

> **⚠️ Experimental SDK**  
> This SDK is experimental and may be discontinued or never released. Use at your own risk.

## Installing the dependency

```bash
go get github.com/bavix/gripmock/v3/pkg/sdk@sdk
```

For local development with gripmock next to this repo, add to go.mod:

```
replace github.com/bavix/gripmock/v3 => ../gripmock
```

Then run `go mod tidy`.

## Generating helloworld

```bash
make generate
```

This generates `helloworld/*.pb.go` from `service.proto` via protoc.

## Running tests

```bash
make test
# or
go test -v ./...
```
