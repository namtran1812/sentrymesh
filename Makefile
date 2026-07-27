.PHONY: run test fmt

run:
	cd gateway && go run ./cmd/sentrymesh

test:
	cd gateway && go test ./...

fmt:
	cd gateway && gofmt -w .
