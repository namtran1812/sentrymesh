package auth

var DefaultStore Repository

func SetDefaultStore(
	store Repository,
) {
	if store == nil {
		panic(
			"auth repository cannot be nil",
		)
	}

	DefaultStore = store
}
