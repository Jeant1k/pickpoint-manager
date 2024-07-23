package module

import (
	"context"
	"errors"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/cli"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
)

type Repository interface {
	AddOrder(ctx context.Context, pickPointId models.PickPointId, order models.Order) error
	RemoveOrder(ctx context.Context, pickPointId models.PickPointId, orderId models.OrderId) error
	FindOrder(ctx context.Context, orderId models.OrderId) (models.Order, error)
	IssueOrder(ctx context.Context, pickPointId models.PickPointId, orderId models.OrderId) error
	ListOrders(ctx context.Context, clientId models.ClientId, limit models.Limit, pickPointId models.PickPointId) ([]models.OrderWithPickPoint, error)
	ReturnOrder(ctx context.Context, pickPointId models.PickPointId, clientId models.ClientId, orderId models.OrderId) error
	ListReturns(ctx context.Context, page models.Page, pageSize models.PageSize) ([]models.Order, error)
}

type TransactionMaganer interface {
	RunRepeatebleRead(ctx context.Context, fx func(ctxTX context.Context) error) error
}

type Deps struct {
	Repository  Repository
	PickPointId models.PickPointId
	TransactionMaganer
}

type Module struct {
	Deps
}

func NewModule(d Deps) *Module {
	return &Module{Deps: d}
}

func (m *Module) RegistratePickPointId(ctx context.Context, pickPointId models.PickPointId) {
	span, _ := opentracing.StartSpanFromContext(ctx, "module.RegistratePickPointId")
	defer span.Finish()
	m.PickPointId = pickPointId
}

func (m *Module) AcceptOrderFromCurier(ctx context.Context, order models.Order) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "module.AcceptOrderFromCurier")
	defer span.Finish()

	existingOrder, errFind := m.Deps.Repository.FindOrder(ctx, order.OrderId)
	if errFind != nil {
		return errors.New("[error in FindOrder] " + errFind.Error())
	}
	if err := validateAcceptOrderFromCurier(existingOrder); err != nil {
		return errors.New("[error in validateAcceptOrderFromCurier] " + err.Error())
	}

	if err := m.Deps.Repository.AddOrder(ctx, m.Deps.PickPointId, order); err != nil {
		return errors.New("[error in AddOrder] " + err.Error())
	}

	return nil
}

func validateAcceptOrderFromCurier(order models.Order) error {
	if order.OrderId != 0 {
		return errors.New("заказ уже существует")
	}
	return nil
}

func (m *Module) ReturnOrderToCurier(ctx context.Context, orderId models.OrderId) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "module.ReturnOrderToCurier")
	defer span.Finish()

	existingOrder, errFind := m.Deps.Repository.FindOrder(ctx, orderId)
	if errFind != nil {
		return errFind
	}
	if err := validateReturnOrderToCurier(existingOrder); err != nil {
		return err
	}

	return m.Repository.RemoveOrder(ctx, m.Deps.PickPointId, orderId)
}

func validateReturnOrderToCurier(order models.Order) error {
	if order.OrderId == 0 {
		return errors.New("заказ не найден")
	}
	if order.Issued {
		return errors.New("заказ был выдан клиенту")
	}
	if order.ShelfLife.After(time.Now().UTC()) {
		return errors.New("shelfLife в будущем")
	}
	if order.Deleted {
		return errors.New("заказ был возвращен курьеру")
	}
	return nil
}

func (m *Module) IssueOrderToClient(ctx context.Context, orderIds map[models.OrderId]bool) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "module.IssueOrderToClient")
	defer span.Finish()

	errIssue := m.TransactionMaganer.RunRepeatebleRead(ctx, func(ctxTX context.Context) error {
		var clients []models.ClientId

		for orderId := range orderIds {
			order, errFind := m.Repository.FindOrder(ctxTX, orderId)
			if errFind != nil {
				return errFind
			}
			if err := validateIssueOrderToClient(order); err != nil {
				return err
			}

			clients = append(clients, order.ClientId)
		}

		clientId := clients[0]
		for _, client := range clients {
			if client != clientId {
				return errors.New("заказы имеют различные clientId")
			}
		}

		for orderId := range orderIds {
			errIssue := m.Repository.IssueOrder(ctxTX, m.Deps.PickPointId, orderId)
			if errIssue != nil {
				return errIssue
			}
		}

		return nil
	})

	return errIssue
}

func validateIssueOrderToClient(order models.Order) error {
	if order.OrderId == 0 {
		return errors.New("заказ не найден")
	}
	if order.Issued {
		return errors.New("заказ был выдан клиенту")
	}
	if order.Returned {
		return errors.New("заказ был возвращен клиентом")
	}
	if order.Deleted {
		return errors.New("заказ был возвращен курьеру")
	}
	if order.ShelfLife.Before(time.Now().UTC()) {
		return errors.New("shelfLife в прошлом")
	}
	return nil
}

func (m *Module) ListOrders(ctx context.Context, clientId models.ClientId, limit models.Limit, pickPointId models.PickPointId) ([]models.OrderWithPickPoint, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "module.ListOrders")
	defer span.Finish()

	return m.Repository.ListOrders(ctx, clientId, limit, pickPointId)
}

func (m *Module) AcceptReturnFromClient(ctx context.Context, clientId models.ClientId, orderId models.OrderId) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "module.AcceptReturnFromClient")
	defer span.Finish()

	existingOrder, errFind := m.Deps.Repository.FindOrder(ctx, orderId)
	if errFind != nil {
		return errFind
	}
	if err := validateAcceptReturnFromClient(existingOrder); err != nil {
		return err
	}

	return m.Repository.ReturnOrder(ctx, m.Deps.PickPointId, clientId, orderId)
}

func validateAcceptReturnFromClient(order models.Order) error {
	if order.OrderId == 0 {
		return errors.New("заказ не найден")
	}
	if !order.Issued {
		return errors.New("заказ не был выдан клиенту")
	}
	if order.Returned {
		return errors.New("заказ был возвращен клиентом")
	}
	if order.Deleted {
		return errors.New("заказ был возвращен курьеру")
	}
	if order.IssueDate.Before(time.Now().UTC().Add(-cli.ReturnTime * 24 * time.Hour)) {
		return errors.New("время возврата просрочено")
	}
	return nil
}

func (m *Module) ListReturns(ctx context.Context, page models.Page, pageSize models.PageSize) ([]models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "module.ListReturns")
	defer span.Finish()

	return m.Repository.ListReturns(ctx, page, pageSize)
}
