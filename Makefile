.PHONY: run test fmt eval integration

AUDIT_DB := $(CURDIR)/sentrymesh-audit.db
APPROVAL_DB := $(CURDIR)/sentrymesh-approvals.db
AUTH_DB := $(CURDIR)/sentrymesh-auth.db
ABUSE_DB := $(CURDIR)/sentrymesh-abuse.db

run:
	cd gateway && go build -o /tmp/sentrymesh ./cmd/sentrymesh
	SENTRYMESH_ROOT="$(CURDIR)" \
	SENTRYMESH_AUDIT_DB="$(AUDIT_DB)" \
	SENTRYMESH_APPROVAL_DB="$(APPROVAL_DB)" \
	SENTRYMESH_AUTH_DB="$(AUTH_DB)" \
	SENTRYMESH_ABUSE_DB="$(ABUSE_DB)" \
	/tmp/sentrymesh

test:
	cd gateway && go test ./...

fmt:
	cd gateway && gofmt -w .

eval:
	cd gateway && SENTRYMESH_ROOT="$(CURDIR)" go run ./cmd/eval

integration:
	cd gateway && go test -v ./integration
