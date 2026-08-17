package api

import "github.com/namtran1812/sentrymesh/gateway/internal/abuse"

var abuseStore abuse.Repository

func SetAbuseStore(
	store abuse.Repository,
) {
	if store == nil {
		panic(
			"abuse repository cannot be nil",
		)
	}

	abuseStore = store
}
