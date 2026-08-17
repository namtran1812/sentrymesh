.PHONY: run test race fmt eval integration integration-postgres

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

race:
	cd gateway && go test -race ./...

fmt:
	cd gateway && gofmt -w .

eval:
	cd gateway && SENTRYMESH_ROOT="$(CURDIR)" go run ./cmd/eval

integration:
	unset DATABASE_URL; \
	SENTRYMESH_AUTH_DB="$$(mktemp -u /tmp/sentrymesh-auth.XXXXXX.db)" \
	SENTRYMESH_AUDIT_DB="$$(mktemp -u /tmp/sentrymesh-audit.XXXXXX.db)" \
	SENTRYMESH_APPROVAL_DB="$$(mktemp -u /tmp/sentrymesh-approval.XXXXXX.db)" \
	SENTRYMESH_ABUSE_DB="$$(mktemp -u /tmp/sentrymesh-abuse.XXXXXX.db)" \
	./scripts/integration.sh

integration-postgres:
	DATABASE_URL="$${DATABASE_URL:-postgresql://sentrymesh:sentrymesh@localhost:5432/sentrymesh}" \
	./scripts/integration.sh
