//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/repository"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/repository/transactor"
)

func TestFindOrder(t *testing.T) {
	// t.Parallel()
	var (
		ctx = context.Background()
	)

	// arrange
	db.SetUp(t, "orders")
	defer db.TearDown(t)
	defer db.TruncateTable(ctx, "orders")
	fillDb(t, order1, pickPointId1)
	transactionManager := transactor.NewTransactionManager(db.DB)
	repo := repository.NewRepository(transactionManager)

	// act
	resp, err := repo.FindOrder(ctx, models.OrderId(order1.OrderId))

	// assert
	require.NoError(t, err)
	assert.Equal(t, order1.OrderId, resp.OrderId)
	assert.Equal(t, order1.ClientId, resp.ClientId)
}

func TestFindOrder_NotFound(t *testing.T) {
	// t.Parallel()
	var (
		ctx     = context.Background()
		orderId = models.OrderId(123)
	)
	db.SetUp(t, "orders")
	defer db.TearDown(t)
	defer db.TruncateTable(ctx, "orders")
	transactionManager := transactor.NewTransactionManager(db.DB)
	repo := repository.NewRepository(transactionManager)

	// act
	resp, _ := repo.FindOrder(ctx, orderId)

	// assert
	assert.Equal(t, models.Order{}, resp)
}

func fillDb(t *testing.T, order models.Order, pickPointId models.PickPointId) {
	_, err := db.DB.Exec(context.Background(), `
		INSERT INTO orders (
			order_id, pick_point_id, client_id, added_date, shelf_life, issued, issue_date, returned, return_date, deleted, delete_date, order_hash
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`,
		order.OrderId,
		pickPointId,
		order.ClientId,
		order.AddedDate.Format(time.RFC3339),
		order.ShelfLife.Format(time.RFC3339),
		order.Issued,
		order.IssueDate.Format(time.RFC3339),
		order.Returned,
		order.ReturnDate.Format(time.RFC3339),
		order.Deleted,
		order.DeleteDate.Format(time.RFC3339),
		order.OrderHash,
	)
	require.NoError(t, err)
}

func TestAddOrder(t *testing.T) {
	var (
		ctx = context.Background()
	)

	// arrange
	db.SetUp(t, "orders")
	defer db.TearDown(t)
	defer db.TruncateTable(ctx, "orders")
	transactionManager := transactor.NewTransactionManager(db.DB)
	repo := repository.NewRepository(transactionManager)

	// act
	err := repo.AddOrder(ctx, pickPointId1, order1)

	// assert
	require.NoError(t, err)

	resp, err := repo.FindOrder(ctx, models.OrderId(order1.OrderId))
	require.NoError(t, err)
	assert.Equal(t, order1.OrderId, resp.OrderId)
	assert.Equal(t, order1.ClientId, resp.ClientId)
	assert.Equal(t, order1.AddedDate.Format(time.RFC3339), resp.AddedDate.Format(time.RFC3339))
	assert.Equal(t, order1.ShelfLife.Format(time.RFC3339), resp.ShelfLife.Format(time.RFC3339))
	assert.Equal(t, order1.OrderHash, resp.OrderHash)
	assert.Equal(t, order1.Weight, resp.Weight)
	assert.Equal(t, order1.Cost, resp.Cost)
	assert.Equal(t, order1.Package.Type(), resp.Package.Type())
	assert.Equal(t, order1.Package.Cost(), resp.Package.Cost())
}

func TestRemoveOrder(t *testing.T) {
	var (
		ctx = context.Background()
	)

	// arrange
	db.SetUp(t, "orders")
	defer db.TearDown(t)
	defer db.TruncateTable(ctx, "orders")
	transactionManager := transactor.NewTransactionManager(db.DB)
	repo := repository.NewRepository(transactionManager)

	err := repo.AddOrder(ctx, pickPointId1, order1)
	require.NoError(t, err)

	// act
	err = repo.RemoveOrder(ctx, pickPointId1, order1.OrderId)

	// assert
	require.NoError(t, err)

	resp, err := repo.FindOrder(ctx, models.OrderId(order1.OrderId))
	require.NoError(t, err)
	assert.Equal(t, order1.OrderId, resp.OrderId)
	assert.Equal(t, order1.ClientId, resp.ClientId)
	assert.Equal(t, order1.AddedDate.Format(time.RFC3339), resp.AddedDate.Format(time.RFC3339))
	assert.Equal(t, order1.ShelfLife.Format(time.RFC3339), resp.ShelfLife.Format(time.RFC3339))
	assert.Equal(t, order1.OrderHash, resp.OrderHash)
	assert.Equal(t, order1.Weight, resp.Weight)
	assert.Equal(t, order1.Cost, resp.Cost)
	assert.Equal(t, order1.Package.Type(), resp.Package.Type())
	assert.Equal(t, order1.Package.Cost(), resp.Package.Cost())
	assert.True(t, resp.Deleted)
}

func TestIssueOrder(t *testing.T) {
	var (
		ctx = context.Background()
	)

	// arrange
	db.SetUp(t, "orders")
	defer db.TearDown(t)
	defer db.TruncateTable(ctx, "orders")
	transactionManager := transactor.NewTransactionManager(db.DB)
	repo := repository.NewRepository(transactionManager)

	err := repo.AddOrder(ctx, pickPointId1, order1)
	require.NoError(t, err)

	// act
	err = repo.IssueOrder(ctx, pickPointId1, order1.OrderId)

	// assert
	require.NoError(t, err)

	resp, err := repo.FindOrder(ctx, models.OrderId(order1.OrderId))
	require.NoError(t, err)
	assert.Equal(t, order1.OrderId, resp.OrderId)
	assert.Equal(t, order1.ClientId, resp.ClientId)
	assert.Equal(t, order1.AddedDate.Format(time.RFC3339), resp.AddedDate.Format(time.RFC3339))
	assert.Equal(t, order1.ShelfLife.Format(time.RFC3339), resp.ShelfLife.Format(time.RFC3339))
	assert.Equal(t, order1.OrderHash, resp.OrderHash)
	assert.Equal(t, order1.Weight, resp.Weight)
	assert.Equal(t, order1.Cost, resp.Cost)
	assert.Equal(t, order1.Package.Type(), resp.Package.Type())
	assert.Equal(t, order1.Package.Cost(), resp.Package.Cost())
	assert.True(t, resp.Issued)
}

func TestReturnOrder(t *testing.T) {
	var (
		ctx = context.Background()
	)

	// arrange
	db.SetUp(t, "orders")
	defer db.TearDown(t)
	defer db.TruncateTable(ctx, "orders")
	transactionManager := transactor.NewTransactionManager(db.DB)
	repo := repository.NewRepository(transactionManager)

	err := repo.AddOrder(ctx, pickPointId1, order1)
	require.NoError(t, err)

	// act
	err = repo.ReturnOrder(ctx, pickPointId1, order1.ClientId, order1.OrderId)

	// assert
	require.NoError(t, err)

	resp, err := repo.FindOrder(ctx, models.OrderId(order1.OrderId))
	require.NoError(t, err)
	assert.Equal(t, order1.OrderId, resp.OrderId)
	assert.Equal(t, order1.ClientId, resp.ClientId)
	assert.Equal(t, order1.AddedDate.Format(time.RFC3339), resp.AddedDate.Format(time.RFC3339))
	assert.Equal(t, order1.ShelfLife.Format(time.RFC3339), resp.ShelfLife.Format(time.RFC3339))
	assert.Equal(t, order1.OrderHash, resp.OrderHash)
	assert.Equal(t, order1.Weight, resp.Weight)
	assert.Equal(t, order1.Cost, resp.Cost)
	assert.Equal(t, order1.Package.Type(), resp.Package.Type())
	assert.Equal(t, order1.Package.Cost(), resp.Package.Cost())
	assert.True(t, resp.Returned)
}
