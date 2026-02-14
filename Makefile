.PHONY: generate test lint lint-fix fmt

GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0

generate:
	protoc --go_out=. --go_opt=module=github.com/bavix/greeter-gripmock-embedded \
		--go-grpc_out=. --go-grpc_opt=module=github.com/bavix/greeter-gripmock-embedded \
		service.proto

test: generate
	go test -race -cover ./...

lint:
	$(GOLANGCI_LINT) run --color always

lint-fix:
	$(GOLANGCI_LINT) run --color always --fix

fmt:
	go fmt ./...
