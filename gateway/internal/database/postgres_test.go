package database

import (
	"context"
	"testing"
)

func TestInvalidPostgresURL(
	t *testing.T,
) {
	_, err := OpenPostgres(
		context.Background(),
		"://invalid",
	)

	if err == nil {
		t.Fatal(
			"expected invalid database URL to fail",
		)
	}
}
