package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/color"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/kafka"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/metrics"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/outbox"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/pkg/api/proto/pickpoint/v1/pickpoint/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Module interface {
	AcceptOrderFromCurier(ctx context.Context, order models.Order) error
	ReturnOrderToCurier(ctx context.Context, orderId models.OrderId) error
	IssueOrderToClient(ctx context.Context, orderIds map[models.OrderId]bool) error
	ListOrders(ctx context.Context, clientId models.ClientId, limit models.Limit, pickPointId models.PickPointId) ([]models.OrderWithPickPoint, error)
	AcceptReturnFromClient(ctx context.Context, clientId models.ClientId, orderId models.OrderId) error
	ListReturns(ctx context.Context, page models.Page, pageSize models.PageSize) ([]models.Order, error)
	RegistratePickPointId(ctx context.Context, pickPointId models.PickPointId)
}

type Deps struct {
	Module      Module
	PickPointId models.PickPointId
	Outbox      *outbox.Outbox
	Producer    *kafka.Producer
	Consumer    *kafka.Consumer
	Ctx         context.Context
}

type PickPointService struct {
	Deps
	commandList []Command
	pickpoint.UnimplementedPickpointServer
}

// NewPickPointService создает интерфейс сервиса
func NewPickPointService(d Deps) PickPointService {
	pickPointService := PickPointService{
		Deps: d,
		commandList: []Command{
			{
				Name:        Help,
				Description: "вывести справку",
			},
			{
				Name:        AcceptOrderFromCurier,
				Description: "принять заказ от курьера: использование acceptCurier --orderId=123 --clientId=123 --shelfLife=2024-08-20T22:42:00+03:00 --weight=12.3 --cost=12.3 --packageType=bag/box/film",
			},
			{
				Name:        ReturnOrderToCurier,
				Description: "вернуть заказ курьеру: использование returnCurier --orderId=123",
			},
			{
				Name:        IssueOrderToClient,
				Description: "выдать заказ клиенту: использование issueClient --orderIds=123,124,125",
			},
			{
				Name:        ListOrders,
				Description: "получить список заказов: использование listOrders --clientId=123 [--limit=10 --pickPointId=123]",
			},
			{
				Name:        AcceptReturnFromClient,
				Description: "принять возврат от клиента: использование acceptClient --clientId=123 --orderId=123",
			},
			{
				Name:        ListReturns,
				Description: "получить список возвратов: использование listReturns --page=1 --pageSize=10",
			},
			{
				Name:        Exit,
				Description: "Закрыть программу",
			},
		},
	}

	return pickPointService
}

// RegistratePickPointId получает от пользователя PickPointId и логиниться на сервере при помощи него.
// Вся дальнейшая работа пользователя с сервисом происходит в контексте этого PickPointId.
func (p *PickPointService) RegistratePickPointId(_ context.Context, req *pickpoint.RegistratePickPointIdRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(p.Ctx, "service.RegistratePickPointId")
	p.Ctx = ctx
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, errors.New("[error in ValidateAll] " + err.Error())
	}

	p.PickPointId = models.PickPointId(req.GetPickPointId().GetPickPointId())
	p.Module.RegistratePickPointId(p.Ctx, p.PickPointId)

	p.logRequest(p.Ctx, "RegistratePickPointId", req)

	return &emptypb.Empty{}, nil
}

// acceptOrderFromCurier принимает заказ от курьера. Она на вход принимает ID заказа, ID получателя
// и срок хранения. Заказ нельзя принять дважды. Если срок хранения в прошлом, выдает ошибку. Функция
// если нужно создает файл, записывает в него для нужного PickPointId информацию о заказе.
func (p *PickPointService) AcceptOrderFromCurier(_ context.Context, req *pickpoint.AcceptOrderFromCurierRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(p.Ctx, "service.AcceptOrderFromCurier")
	p.Ctx = ctx
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, errors.New("[error in ValidateAll] " + err.Error())
	}

	order, errToDomain := orderToDomain(req)
	if errToDomain != nil {
		return nil, errors.New("[error in orderToDomain] " + errToDomain.Error())
	}

	errAcceptOrder := p.Module.AcceptOrderFromCurier(p.Ctx, order)
	if errAcceptOrder == nil {
		p.logRequest(p.Ctx, "acceptOrderFromCurier", req)
		return &emptypb.Empty{}, nil
	}

	return nil, errors.New("[error in AcceptOrderFromCurier] " + errAcceptOrder.Error())
}

func orderToDomain(req *pickpoint.AcceptOrderFromCurierRequest) (models.Order, error) {
	packageType, errConv := packageTypeToString(req.GetOrder().GetPackage().Enum())
	if errConv != nil {
		return models.Order{}, errors.New("[error in packageTypeToString] " + errConv.Error())
	}
	pkg, errBuild := models.PackageFactory(packageType)
	if errBuild != nil {
		return models.Order{}, errors.New("[error in PackageFactory] " + errBuild.Error())
	}

	order := models.Order{
		OrderId:    models.OrderId(req.GetOrder().GetOrderId().GetOrderId()),
		ClientId:   models.ClientId(req.GetOrder().GetClientId().GetClientId()),
		AddedDate:  models.NewAddedDate(req.GetOrder().GetAddedDate().AsTime()),
		ShelfLife:  models.NewShelfLife(req.GetOrder().GetShelfLife().AsTime()),
		Issued:     req.GetOrder().GetIssued(),
		IssueDate:  models.NewIssueDate(req.GetOrder().GetIssueDate().AsTime()),
		Returned:   req.GetOrder().GetReturned(),
		ReturnDate: models.NewReturnDate(req.GetOrder().GetReturnDate().AsTime()),
		Deleted:    req.GetOrder().GetDeleted(),
		DeleteDate: models.NewDeleteDate(req.GetOrder().GetDeletedDate().AsTime()),
		OrderHash:  models.Hash(req.GetOrder().GetHash()),
		Weight:     models.Weight(req.GetOrder().GetWeight()),
		Cost:       models.Cost(req.GetOrder().GetCost()),
		Package:    pkg,
	}
	if err := pkg.Validate(order); err != nil {
		return models.Order{}, errors.New("[error in Validate] " + err.Error())
	}

	return order, nil
}

// returnOrderToCurier возвращает заказ курьеру. На вход принимается ID заказа. Функция помечает
// заказ в файле удаленным. Можно вернуть только те заказы, у которых вышел срок хранения и если
// заказы не были выданы клиенту. Также проверяется находится ли заказ в текущем ПВЗ, чтобы нельзя
// было вернуть заказ курьеру из другого ПВЗ.
func (p *PickPointService) ReturnOrderToCurier(_ context.Context, req *pickpoint.ReturnOrderToCurierRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(p.Ctx, "service.ReturnOrderToCurier")
	p.Ctx = ctx
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, errors.New("[error in ValidateAll] " + err.Error())
	}

	orderId := models.OrderId(req.GetOrderId().GetOrderId())

	errReturnOrder := p.Module.ReturnOrderToCurier(p.Ctx, orderId)

	if errReturnOrder == nil {
		p.logRequest(p.Ctx, "returnOrderToCurier", req)
		return &emptypb.Empty{}, nil
	}

	return nil, errors.New("[error in ReturnOrderToCurier] " + errReturnOrder.Error())
}

// issueOrderToClient выдает заказ клиенту. На вход принимается список ID заказов. Можно выдавать
// только те заказы, которые были приняты от курьера и чей срок хранения меньше текущей даты. Все ID
// заказов должны принадлежать только одному клиенту. Также проверяется находится ли заказ(ы) в текущем
// ПВЗ, чтобы нельзя было выдать заказ(ы) клиенту из другого ПВЗ.
func (p *PickPointService) IssueOrderToClient(_ context.Context, req *pickpoint.IssueOrderToClientRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(p.Ctx, "service.IssueOrderToClient")
	p.Ctx = ctx
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, errors.New("[error in ValidateAll] " + err.Error())
	}

	orderIdMap := make(map[models.OrderId]bool)
	for _, orderId := range req.GetOrderIds() {
		orderIdMap[models.OrderId(orderId.GetOrderId())] = true
	}

	errIssueOrder := p.Module.IssueOrderToClient(p.Ctx, orderIdMap)

	if errIssueOrder == nil {
		p.logRequest(p.Ctx, "issueOrderToClient", req)

		for orderId := range orderIdMap {
			metrics.IncNumberIssuedOrders(fmt.Sprint(orderId), fmt.Sprint(p.PickPointId))
		}
		return &emptypb.Empty{}, nil
	}

	return nil, errors.New("[error in IssueOrderToClient] " + errIssueOrder.Error())
}

// listOrders возаращает список заказов, которые не были выданы, возвращены или удалены. На вход
// принимается ID пользователя как обязательный параметр и опциональные параметры. Параметры позволяют
// получать только последние N заказов или заказы клиента, находящиеся в нашем ПВЗ.
func (p *PickPointService) ListOrders(_ context.Context, req *pickpoint.ListOrdersRequest) (*pickpoint.ListOrdersResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(p.Ctx, "service.ListOrders")
	p.Ctx = ctx
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, errors.New("[error in ValidateAll] " + err.Error())
	}

	clientId, limit, pickPointId := models.ClientId(req.GetClientId().GetClientId()), models.Limit(req.GetLimit()), models.PickPointId(req.GetPickPointId().GetPickPointId())

	orders, errListOrders := p.Module.ListOrders(p.Ctx, clientId, limit, pickPointId)
	if errListOrders != nil {
		return nil, errListOrders
	}

	var list []*pickpoint.OrderWithPickPoint
	for _, order := range orders {
		list = append(list, &pickpoint.OrderWithPickPoint{
			Order: domainToOrder(order.Order),
			PickPointId: &pickpoint.PickPointId{
				PickPointId: int64(order.PickPointId),
			},
		})
	}

	p.logRequest(p.Ctx, "listOrders", req)

	return &pickpoint.ListOrdersResponse{
		List: list,
	}, nil
}

// acceptReturnFromClient принимает возврат от клиента. На вход принимается ID пользователя и ID заказа.
// Заказ может быть возвращен в течение двух дней с момента выдачи. Также она проверяет, что заказ
// выдавался с нашего ПВЗ.
func (p *PickPointService) AcceptReturnFromClient(_ context.Context, req *pickpoint.AcceptReturnFromClientRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(p.Ctx, "service.AcceptReturnFromClient")
	p.Ctx = ctx
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, errors.New("[error in ValidateAll] " + err.Error())
	}

	clientId, orderId := models.ClientId(req.GetClientId().GetClientId()), models.OrderId(req.GetOrderId().GetOrderId())

	errAcceptReturn := p.Module.AcceptReturnFromClient(p.Ctx, clientId, orderId)

	if errAcceptReturn == nil {
		fmt.Println("заказ с orderId =", orderId, "возвращен от клиента")

		p.logRequest(p.Ctx, "acceptReturnFromClient", req)
		return &emptypb.Empty{}, nil
	}

	return nil, errors.New("[error in AcceptReturnFromClient] " + errAcceptReturn.Error())
}

// listReturns возвращает список заказов. Функция выдает список пагинированно.
func (p *PickPointService) ListReturns(_ context.Context, req *pickpoint.ListReturnsRequest) (*pickpoint.ListReturnsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(p.Ctx, "service.ListReturns")
	p.Ctx = ctx
	defer span.Finish()

	if err := req.ValidateAll(); err != nil {
		return nil, errors.New("[error in ValidateAll] " + err.Error())
	}

	page, pageSize := models.Page(req.GetPage()), models.PageSize(req.GetPageSize())

	returns, errListReturns := p.Module.ListReturns(p.Ctx, page, pageSize)
	if errListReturns != nil {
		return nil, errListReturns
	}

	var list []*pickpoint.Order
	for _, order := range returns {
		list = append(list, domainToOrder(order))
	}

	p.logRequest(p.Ctx, "listReturns", req)

	return &pickpoint.ListReturnsResponse{
		Orders: list,
	}, nil
}

func domainToOrder(order models.Order) *pickpoint.Order {
	return &pickpoint.Order{
		OrderId:     &pickpoint.OrderId{OrderId: int64(order.OrderId)},
		ClientId:    &pickpoint.ClientId{ClientId: int64(order.ClientId)},
		AddedDate:   timestamppb.New(order.AddedDate.Time),
		ShelfLife:   timestamppb.New(order.ShelfLife.Time),
		Issued:      order.Issued,
		IssueDate:   timestamppb.New(order.IssueDate.Time),
		Returned:    order.Returned,
		ReturnDate:  timestamppb.New(order.ReturnDate.Time),
		Deleted:     order.Deleted,
		DeletedDate: timestamppb.New(order.DeleteDate.Time),
		Hash:        string(order.OrderHash),
	}
}

// help выводит справку
func (p *PickPointService) Help(context.Context, *emptypb.Empty) (*pickpoint.HelpResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(p.Ctx, "service.Help")
	p.Ctx = ctx
	defer span.Finish()

	var commands []*pickpoint.Command
	for _, cmd := range p.commandList {
		commands = append(commands, &pickpoint.Command{
			Name:        cmd.Name,
			Description: cmd.Description,
		})
	}

	p.logRequest(p.Ctx, "help", []string{})

	return &pickpoint.HelpResponse{Commands: commands}, nil
}

func packageTypeToString(packageType *pickpoint.Order_PackageType) (string, error) {
	switch *packageType {
	case pickpoint.Order_PACKAGE_TYPE_BAG:
		return "bag", nil
	case pickpoint.Order_PACKAGE_TYPE_BOX:
		return "box", nil
	case pickpoint.Order_PACKAGE_TYPE_FILM:
		return "film", nil
	default:
		return "", errors.New("неизвестный тип упаковки")
	}
}

func (p *PickPointService) logRequest(ctx context.Context, methodName string, req interface{}) {
	event := map[string]interface{}{
		"time":  time.Now(),
		"event": methodName,
		"query": req,
	}
	eventJSON, _ := json.Marshal(event)

	if p.Deps.Producer != nil {
		err := p.Deps.Producer.SendMessage("logs", string(eventJSON))
		if err != nil {
			color.PrintRed("\tОшибка отправки события в Kafka: " + err.Error())
			p.Deps.Outbox.AddEvent(ctx, methodName, string(eventJSON))
		}
	} else {
		color.PrintYellow("logs " + string(eventJSON))
	}
}
