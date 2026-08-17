package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/namtran1812/sentrymesh/gateway/internal/abuse"
	"github.com/namtran1812/sentrymesh/gateway/internal/approval"
	"github.com/namtran1812/sentrymesh/gateway/internal/audit"
	"github.com/namtran1812/sentrymesh/gateway/internal/auth"
)

func TestMain(m *testing.M) {
	root, err := os.MkdirTemp(
		"",
		"sentrymesh-api-tests-*",
	)
	if err != nil {
		panic(err)
	}

	authStore, err := auth.NewStore(
		filepath.Join(root, "auth.db"),
	)
	if err != nil {
		panic(err)
	}

	approvalStore, err := approval.NewStore(
		filepath.Join(root, "approvals.db"),
	)
	if err != nil {
		_ = authStore.Close()
		panic(err)
	}

	auditStore, err := audit.NewStore(
		filepath.Join(root, "audit.db"),
	)
	if err != nil {
		_ = approvalStore.Close()
		_ = authStore.Close()
		panic(err)
	}

	if err := auditStore.EnsureToolEvents(); err != nil {
		panic(err)
	}

	if err := auditStore.EnsureAuthEvents(); err != nil {
		panic(err)
	}

	if err := auditStore.EnsureRAGEvents(); err != nil {
		panic(err)
	}

	if err := auditStore.EnsureAbuseEvents(); err != nil {
		panic(err)
	}

	abuseStore, err := abuse.NewStore(
		filepath.Join(root, "abuse.db"),
	)
	if err != nil {
		panic(err)
	}

	auth.SetDefaultStore(authStore)
	SetApprovalStore(approvalStore)
	SetAuditStore(auditStore)
	SetAbuseStore(abuseStore)

	code := m.Run()

	_ = abuseStore.Close()
	_ = auditStore.Close()
	_ = approvalStore.Close()
	_ = authStore.Close()
	_ = os.RemoveAll(root)

	os.Exit(code)
}
