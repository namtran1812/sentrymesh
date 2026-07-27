.PHONY: run test fmt

AUDIT_DB := $(CURDIR)/sentrymesh-audit.db
APPROVAL_DB := $(CURDIR)/sentrymesh-approvals.db
AUTH_DB := $(CURDIR)/sentrymesh-auth.db

run:
	cd gateway && \
	SENTRYMESH_AUDIT_DB="$(AUDIT_DB)" \
	SENTRYMESH_APPROVAL_DB="$(APPROVAL_DB)" \
	SENTRYMESH_AUTH_DB="$(AUTH_DB)" \
	go run ./cmd/sentrymesh

test:
	cd gateway && go test ./...

fmt:
	cd gateway && gofmt -w .
