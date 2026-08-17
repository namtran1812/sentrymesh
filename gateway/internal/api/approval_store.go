package api

import "github.com/namtran1812/sentrymesh/gateway/internal/approval"

var approvalStore approval.Repository

func SetApprovalStore(
	store approval.Repository,
) {
	if store == nil {
		panic(
			"approval repository cannot be nil",
		)
	}

	approvalStore = store
}
