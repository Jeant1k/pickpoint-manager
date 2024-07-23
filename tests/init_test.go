//go:build integration

package tests

import "gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/tests/postgresql"

var (
	db *postgresql.TDB
)

func init() {
	db = postgresql.New()
}
