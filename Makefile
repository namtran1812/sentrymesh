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
	@tmpdir=$$(mktemp -d); \
	trap 'rm -rf "$$tmpdir"' EXIT; \
	unset DATABASE_URL; \
	SENTRYMESH_AUTH_DB="$$tmpdir/auth.db" \
	SENTRYMESH_AUDIT_DB="$$tmpdir/audit.db" \
	SENTRYMESH_APPROVAL_DB="$$tmpdir/approval.db" \
	SENTRYMESH_ABUSE_DB="$$tmpdir/abuse.db" \
	./scripts/integration.sh

integration-postgres:
	DATABASE_URL="$${DATABASE_URL:-postgresql://sentrymesh:sentrymesh@localhost:5432/sentrymesh}" \
	./scripts/integration.sh
