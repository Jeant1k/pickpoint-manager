package postgresql

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
)

type TDB struct {
	DB *pgxpool.Pool
}

func New() *TDB {
	dbURL := "postgres://postgres:examplepassword@localhost:5432/oms?sslmode=disable"

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dbURL)
	if err != nil {
		panic(fmt.Sprintf("Ошибка подключения к базе данных: %v", err))
	}

	return &TDB{DB: pool}
}

func (d *TDB) SetUp(t *testing.T, tableName ...string) {
	t.Helper()
	d.TruncateTable(context.Background(), tableName...)
}

func (d *TDB) TearDown(t *testing.T) {
	t.Helper()
	d.DB.Close()
}

func (d *TDB) TruncateTable(ctx context.Context, tableName ...string) {
	q := fmt.Sprintf("TRUNCATE %s", strings.Join(tableName, ","))
	if _, err := d.DB.Exec(ctx, q); err != nil {
		panic(err)
	}
}
