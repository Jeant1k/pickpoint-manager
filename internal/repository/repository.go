package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/cache"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/repository/schema"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/repository/transactor"
)

type Repository struct {
	transactor.QueryEngineProvider
	cache cache.Cache
}

func NewRepository(provider transactor.QueryEngineProvider, cache cache.Cache) *Repository {
	return &Repository{
		QueryEngineProvider: provider,
		cache:               cache,
	}
}

func (r *Repository) FindOrder(ctx context.Context, orderId models.OrderId) (models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.FindOrder")
	defer span.Finish()

	var chacheOrder models.Order
	cacheKey := fmt.Sprintf("order:%d", orderId)
	if r.cache.Get(ctx, cacheKey, &chacheOrder) {
		return chacheOrder, nil
	}

	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	row, errFind := db.Query(
		ctx,
		`
		SELECT order_id, client_id, added_date, shelf_life, issued, issue_date, returned, return_date, deleted, delete_date, order_hash
		FROM orders
		WHERE order_id = $1
		`,
		orderId,
	)
	if errFind != nil {
		return models.Order{}, errors.New("[error in Query] " + errFind.Error())
	}
	defer row.Close()

	var order schema.Order
	if row.Next() {
		errScan := row.Scan(
			&order.OrderId,
			&order.ClientId,
			&order.AddedDate,
			&order.ShelfLife,
			&order.Issued,
			&order.IssueDate,
			&order.Returned,
			&order.ReturnDate,
			&order.Deleted,
			&order.DeleteDate,
			&order.OrderHash,
		)
		if errScan != nil {
			return models.Order{}, errors.New("[error in Scan] " + errScan.Error())
		}
	}

	if order.OrderId != 0 {
		domainOrder := toDomainOrder(order)
		errCache := r.cache.Set(ctx, cacheKey, domainOrder)
		if errCache != nil {
			log.Printf("failed to cache order %d: %v", orderId, errCache)
		}
	}

	return toDomainOrder(order), nil
}

func (r *Repository) AddOrder(ctx context.Context, pickPointId models.PickPointId, order models.Order) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.AddOrder")
	defer span.Finish()

	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	_, errAdd := db.Query(
		ctx,
		`
		INSERT INTO orders (order_id, pick_point_id, client_id, added_date, shelf_life, order_hash, weight, cost, package_type, package_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
		order.OrderId,
		pickPointId,
		order.ClientId,
		order.AddedDate.Format(time.RFC3339),
		order.ShelfLife.Format(time.RFC3339),
		order.OrderHash,
		order.Weight,
		order.Cost,
		order.Package.Type(),
		order.Package.Cost(),
	)
	if errAdd != nil {
		return errors.New("[error in Query] " + errAdd.Error())
	}

	return nil
}

func (r *Repository) RemoveOrder(ctx context.Context, pickPointId models.PickPointId, orderId models.OrderId) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.RemoveOrder")
	defer span.Finish()

	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	_, errRemove := db.Query(
		ctx,
		`
		UPDATE orders
		SET deleted = TRUE, delete_date = $1
		WHERE pick_point_id = $2 AND order_id = $3
		`,
		time.Now().UTC(),
		orderId,
		pickPointId,
	)
	if errRemove != nil {
		return errRemove
	}

	return nil
}

func (r *Repository) IssueOrder(ctx context.Context, pickPointId models.PickPointId, orderId models.OrderId) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.IssueOrder")
	defer span.Finish()

	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	_, errIssue := db.Query(
		ctx,
		`
		UPDATE orders
		SET issued = TRUE, issue_date = $1
		WHERE pick_point_id = $2 AND order_id = $3
		`,
		time.Now().UTC(),
		pickPointId,
		orderId,
	)
	if errIssue != nil {
		return errIssue
	}

	return nil
}

func (r *Repository) ListOrders(ctx context.Context, clientId models.ClientId, limit models.Limit, pickPointId models.PickPointId) ([]models.OrderWithPickPoint, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListOrders")
	defer span.Finish()

	cacheKey := fmt.Sprintf("orders:%d:%d:%d", clientId, pickPointId, limit)
	var cacheOrders []models.OrderWithPickPoint
	if r.cache.Get(ctx, cacheKey, &cacheOrders) {
		return cacheOrders, nil
	}

	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	rows, errList := db.Query(
		ctx,
		`
		SELECT order_id, pick_point_id, client_id, added_date, shelf_life, issued, issue_date, returned, return_date, deleted, delete_date, order_hash
		FROM orders
		WHERE client_id = $1 AND issued = FALSE AND returned = FALSE AND deleted = FALSE
		ORDER BY added_date
		`,
		clientId,
	)
	if errList != nil {
		return nil, errList
	}

	var orders []models.OrderWithPickPoint
	for rows.Next() {
		var order schema.Order
		errScan := rows.Scan(
			&order.OrderId,
			&order.PickPointId,
			&order.ClientId,
			&order.AddedDate,
			&order.ShelfLife,
			&order.Issued,
			&order.IssueDate,
			&order.Returned,
			&order.ReturnDate,
			&order.Deleted,
			&order.DeleteDate,
			&order.OrderHash,
		)
		if errScan != nil && len(orders) == 0 {
			return []models.OrderWithPickPoint{}, errScan
		}

		if pickPointId == 0 || order.PickPointId == int64(pickPointId) {
			orders = append(orders, models.OrderWithPickPoint{
				Order:       toDomainOrder(order),
				PickPointId: models.PickPointId(order.PickPointId),
			})
		}
	}
	if len(orders) == 0 {
		return nil, errors.New("доступных заказов не найдено")
	}

	if int(limit) > 0 && len(orders) > int(limit) {
		orders = orders[:limit]
	}

	errCache := r.cache.Set(ctx, cacheKey, orders)
	if errCache != nil {
		log.Printf("failed to cache orders for client %d: %v", clientId, errCache)
	}

	return orders, nil
}

func (r *Repository) ReturnOrder(ctx context.Context, pickPointId models.PickPointId, clientId models.ClientId, orderId models.OrderId) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ReturnOrder")
	defer span.Finish()

	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	_, errReturn := db.Query(
		ctx,
		`
		UPDATE orders
		SET returned = TRUE, return_date = $1
		WHERE pick_point_id = $2 AND order_id = $3 AND client_id = $4
		`,
		time.Now().UTC(),
		pickPointId,
		orderId,
		clientId,
	)
	if errReturn != nil {
		return errReturn
	}

	return nil
}

func (r *Repository) ListReturns(ctx context.Context, page models.Page, pageSize models.PageSize) ([]models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListReturns")
	defer span.Finish()

	cacheKey := fmt.Sprintf("returns:%d:%d", page, pageSize)
	var cacheReturns []models.Order
	if r.cache.Get(ctx, cacheKey, &cacheReturns) {
		return cacheReturns, nil
	}

	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	rows, errList := db.Query(
		ctx,
		`
		SELECT order_id, pick_point_id, client_id, added_date, shelf_life, issued, issue_date, returned, return_date, deleted, delete_date, order_hash
		FROM orders
		WHERE returned = TRUE
		ORDER BY added_date
		`,
	)
	if errList != nil {
		return nil, errList
	}

	count := 0
	var returns []models.Order
	for rows.Next() {
		count++
		if (int(page)-1)*int(pageSize)+1 <= count && count <= int(page)*int(pageSize) {
			var order schema.Order
			errScan := rows.Scan(
				&order.OrderId,
				&order.PickPointId,
				&order.ClientId,
				&order.AddedDate,
				&order.ShelfLife,
				&order.Issued,
				&order.IssueDate,
				&order.Returned,
				&order.ReturnDate,
				&order.Deleted,
				&order.DeleteDate,
				&order.OrderHash,
			)
			if errScan != nil && len(returns) == 0 {
				return []models.Order{}, errScan
			}

			returns = append(returns, toDomainOrder(order))
		}
	}
	if len(returns) == 0 {
		return nil, errors.New("возвратов не найдено")
	}

	errCache := r.cache.Set(ctx, cacheKey, returns)
	if errCache != nil {
		log.Printf("failed to cache returns: %v", errCache)
	}

	return returns, nil
}

func toDomainOrder(order schema.Order) models.Order {
	return models.Order{
		OrderId:    models.OrderId(order.OrderId),
		ClientId:   models.ClientId(order.ClientId),
		AddedDate:  models.NewAddedDate(order.AddedDate),
		ShelfLife:  models.NewShelfLife(order.ShelfLife),
		Issued:     order.Issued,
		IssueDate:  models.NewIssueDate(order.IssueDate.Time),
		Returned:   order.Returned,
		ReturnDate: models.NewReturnDate(order.ReturnDate.Time),
		Deleted:    order.Deleted,
		DeleteDate: models.NewDeleteDate(order.DeleteDate.Time),
		OrderHash:  models.Hash(order.OrderHash),
	}
}
