package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	service "gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/api"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/color"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/hash"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/pkg/api/proto/pickpoint/v1/pickpoint/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	target = "localhost:50051"
)

func main() {
	conn, errConn := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errConn != nil {
		log.Fatal(errConn)
		return
	}
	defer conn.Close()

	client := pickpoint.NewPickpointClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("\tПривет! Эта программа по управлению пунктом выдачи.")
	fmt.Println("\tЧтобы войти введи pickPointId своего пункта выдачи.")

	scanner := bufio.NewScanner(os.Stdin)
	var pickPointId int64
	for {
		fmt.Print("> ")

		scanner.Scan()
		var err error
		pickPointId, err = strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			color.PrintRed("\tError: " + err.Error())
		}

		if pickPointId <= 0 {
			color.PrintRed("\tError: pickPointId должен быть натуральным числом")
		} else {
			break
		}
	}

	client.RegistratePickPointId(ctx, &pickpoint.RegistratePickPointIdRequest{
		PickPointId: &pickpoint.PickPointId{
			PickPointId: pickPointId,
		},
	})

	fmt.Println("\tДля того чтобы узнать что программа умеет, введите help.")

	for {
		fmt.Printf("%d: ", pickPointId)
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		args := strings.Fields(line)

		if len(args) == 0 {
			continue
		}

		command := args[0]
		args = args[1:]

		if command == service.Exit {
			break
		}

		switch command {
		case service.Help:
			if err := help(ctx, client); err != nil {
				color.PrintRed("\tError: " + err.Error())
			}
		case service.AcceptOrderFromCurier:
			if err := acceptOrderFromCurier(ctx, args, client); err != nil {
				color.PrintRed("\tError: " + err.Error())
			}
		case service.ReturnOrderToCurier:
			if err := returnOrderToCurier(ctx, args, client); err != nil {
				color.PrintRed("\tError: " + err.Error())
			}
		case service.IssueOrderToClient:
			if err := issueOrderToClient(ctx, args, client); err != nil {
				color.PrintRed("\tError: " + err.Error())
			}
		case service.ListOrders:
			if err := listOrders(ctx, args, client); err != nil {
				color.PrintRed("\tError: " + err.Error())
			}
		case service.AcceptReturnFromClient:
			if err := acceptReturnFromClient(ctx, args, client); err != nil {
				color.PrintRed("\tError: " + err.Error())
			}
		case service.ListReturns:
			if err := listReturns(ctx, args, client); err != nil {
				color.PrintRed("\tError: " + err.Error())
			}
		default:
			color.PrintRed("\tНеизвестная команда: " + command)
		}
	}

	fmt.Println("\tДо скорых встреч!")
}

func listReturns(ctx context.Context, args []string, client pickpoint.PickpointClient) error {
	page, pageSize, errParse := parseListReturnsArgs(args)
	if errParse != nil {
		return errors.New("[error in parseListReturnsArgs] " + errParse.Error())
	}

	returns, errListReturns := client.ListReturns(ctx, &pickpoint.ListReturnsRequest{
		Page:     int64(page),
		PageSize: int64(pageSize),
	})
	if errListReturns != nil {
		return errors.New("[error in ListReturns] " + errListReturns.Error())
	}

	for _, ret := range returns.GetOrders() {
		fmt.Printf(
			"\torderId: %d, clientId: %d, addedDate: %s, returnDate: %s\n",
			ret.GetOrderId().GetOrderId(),
			ret.GetClientId().GetClientId(),
			ret.GetAddedDate().AsTime().Local().Format("2006-01-02 15:04:05"),
			ret.GetReturnDate().AsTime().Local().Format("2006-01-02 15:04:05"),
		)
	}

	return nil
}

func parseListReturnsArgs(args []string) (models.Page, models.PageSize, error) {
	var page, pageSize int

	fs := flag.NewFlagSet(service.ListReturns, flag.ContinueOnError)
	fs.IntVar(&page, "page", 1, "используйте --page=1")
	fs.IntVar(&pageSize, "pageSize", 5, "используйте --pageSize=10")
	if err := fs.Parse(args); err != nil {
		return models.Page(0), models.PageSize(0), err
	}

	if page <= 0 {
		return models.Page(0), models.PageSize(0), errors.New("page должен быть натуральным числом")
	}
	if pageSize <= 0 {
		return models.Page(0), models.PageSize(0), errors.New("pageSize должен быть натуральным числом")
	}

	return models.Page(page), models.PageSize(pageSize), nil
}

func acceptReturnFromClient(ctx context.Context, args []string, client pickpoint.PickpointClient) error {
	clientId, orderId, errParse := parseAcceptReturnFromClientArgs(args)
	if errParse != nil {
		return errors.New("[error in parseAcceptReturnFromClientArgs] " + errParse.Error())
	}

	_, errAcceptReturn := client.AcceptReturnFromClient(ctx, &pickpoint.AcceptReturnFromClientRequest{
		ClientId: &pickpoint.ClientId{ClientId: int64(clientId)},
		OrderId:  &pickpoint.OrderId{OrderId: int64(orderId)},
	})
	if errAcceptReturn != nil {
		return errors.New("[error in AcceptReturnFromClient] " + errAcceptReturn.Error())
	}

	fmt.Println("заказ с orderId =", orderId, "возвращен от клиента")

	return nil
}

func parseAcceptReturnFromClientArgs(args []string) (models.ClientId, models.OrderId, error) {
	var clientId, orderId int

	fs := flag.NewFlagSet(service.AcceptReturnFromClient, flag.ContinueOnError)
	fs.IntVar(&clientId, "clientId", 0, "используйте --clientId=123")
	fs.IntVar(&orderId, "orderId", 0, "используйте --orderId=123")
	if err := fs.Parse(args); err != nil {
		return models.ClientId(0), models.OrderId(0), err
	}

	if orderId <= 0 {
		return models.ClientId(0), models.OrderId(0), errors.New("orderId должен быть натуральным числом")
	}
	if clientId <= 0 {
		return models.ClientId(0), models.OrderId(0), errors.New("clientId должен быть натуральным числом")
	}

	return models.ClientId(clientId), models.OrderId(orderId), nil
}

func listOrders(ctx context.Context, args []string, client pickpoint.PickpointClient) error {
	clientId, limit, pickPointId, errParse := parseListOrdersArgs(args)
	if errParse != nil {
		return errors.New("[error in parseListOrdersArgs] " + errParse.Error())
	}

	limitInt64 := int64(limit)
	orders, errListOrders := client.ListOrders(ctx, &pickpoint.ListOrdersRequest{
		ClientId:    &pickpoint.ClientId{ClientId: int64(clientId)},
		Limit:       &limitInt64,
		PickPointId: &pickpoint.PickPointId{PickPointId: int64(pickPointId)},
	})
	if errListOrders != nil {
		return errors.New("[error in ListOrders] " + errListOrders.Error())
	}

	for _, order := range orders.List {
		fmt.Printf(
			"\tpickPointId: %d, orderId: %d, addedDate: %s, shelfLife: %s\n",
			order.GetPickPointId().GetPickPointId(),
			order.GetOrder().GetOrderId().GetOrderId(),
			order.GetOrder().GetAddedDate().AsTime().Local().Format("2006-01-02 15:04:05"),
			order.GetOrder().GetShelfLife().AsTime().Local().Format("2006-01-02 15:04:05"),
		)
	}

	return nil
}

func parseListOrdersArgs(args []string) (models.ClientId, models.Limit, models.PickPointId, error) {
	var clientId, limit, pickPointId int64

	fs := flag.NewFlagSet(service.ListOrders, flag.ContinueOnError)
	fs.Int64Var(&clientId, "clientId", 0, "используйте --clientId=123")
	fs.Int64Var(&limit, "limit", 0, "используйте --limit=10")
	fs.Int64Var(&pickPointId, "pickPointId", 0, "используйте --pickPointId=123")
	if err := fs.Parse(args); err != nil {
		return models.ClientId(0), models.Limit(0), models.PickPointId(0), err
	}

	if clientId <= 0 {
		return models.ClientId(0), models.Limit(0), models.PickPointId(0), errors.New("clientId должен быть натуральным числом")
	}
	if limit < 0 {
		return models.ClientId(0), models.Limit(0), models.PickPointId(0), errors.New("limit должен быть натуральным числом")
	}
	if pickPointId < 0 {
		return models.ClientId(0), models.Limit(0), models.PickPointId(0), errors.New("pickPointId должен быть натуральным числом")
	}

	return models.ClientId(clientId), models.Limit(limit), models.PickPointId(pickPointId), nil
}

func issueOrderToClient(ctx context.Context, args []string, client pickpoint.PickpointClient) error {
	orderIdMap, errParse := parseIssueOrderToClientArgs(args)
	if errParse != nil {
		return errors.New("[error in parseIssueOrderToClientArgs] " + errParse.Error())
	}

	var list []*pickpoint.OrderId
	for orderId := range orderIdMap {
		list = append(list, &pickpoint.OrderId{
			OrderId: int64(orderId),
		})
	}

	_, errIssueOrder := client.IssueOrderToClient(ctx, &pickpoint.IssueOrderToClientRequest{
		OrderIds: list,
	})
	if errIssueOrder != nil {
		return errors.New("[error in IssueOrderToClient] " + errIssueOrder.Error())
	}

	fmt.Print("\tзаказы с orderIds = ")
	for orderId := range orderIdMap {
		fmt.Print(orderId, ", ")
	}
	fmt.Println("выданы клиенту")

	return nil
}

func parseIssueOrderToClientArgs(args []string) (map[models.OrderId]bool, error) {
	var orderIds string

	fs := flag.NewFlagSet(service.IssueOrderToClient, flag.ContinueOnError)
	fs.StringVar(&orderIds, "orderIds", "", "используйте --orderIds=123,124,125")
	if err := fs.Parse(args); err != nil {
		return map[models.OrderId]bool{}, err
	}

	if len(orderIds) == 0 {
		return map[models.OrderId]bool{}, errors.New("orderIds пусто")
	}

	orderIdList := strings.Split(orderIds, ",")
	orderIdMap := make(map[models.OrderId]bool)
	for _, idStr := range orderIdList {
		id, errParse := strconv.ParseInt(idStr, 10, 64)
		if errParse != nil {
			return map[models.OrderId]bool{}, errParse
		}
		if id <= 0 {
			return map[models.OrderId]bool{}, errors.New("orderId должен быть натуральным числом")
		}

		if _, exists := orderIdMap[models.OrderId(id)]; exists {
			return map[models.OrderId]bool{}, errors.New("дублирование orderId")
		}
		orderIdMap[models.OrderId(id)] = true
	}

	return orderIdMap, nil
}

func returnOrderToCurier(ctx context.Context, args []string, client pickpoint.PickpointClient) error {
	orderId, errParse := parseReturnOrderToCurierArgs(args)
	if errParse != nil {
		return errors.New("[error in parseReturnOrderToCurierArgs] " + errParse.Error())
	}

	_, errReturnOrder := client.ReturnOrderToCurier(ctx, &pickpoint.ReturnOrderToCurierRequest{
		OrderId: &pickpoint.OrderId{
			OrderId: int64(orderId),
		},
	})
	if errReturnOrder != nil {
		return errors.New("[error in ReturnOrderToCurier] " + errReturnOrder.Error())
	}

	fmt.Println("\tзаказ с orderId =", orderId, "возвращён курьеру")

	return nil
}

func parseReturnOrderToCurierArgs(args []string) (models.OrderId, error) {
	var orderId int64

	fs := flag.NewFlagSet(service.ReturnOrderToCurier, flag.ContinueOnError)
	fs.Int64Var(&orderId, "orderId", 0, "используйте --orderId=123")
	if err := fs.Parse(args); err != nil {
		return models.OrderId(0), err
	}

	if orderId <= 0 {
		return models.OrderId(0), errors.New("orderId должен быть натуральным числом")
	}

	return models.OrderId(orderId), nil
}

func acceptOrderFromCurier(ctx context.Context, args []string, client pickpoint.PickpointClient) error {
	order, errParse := parseAcceptOrderFromCurierArgs(args)
	if errParse != nil {
		return errors.New("[error in parseAcceptOrderFromCurierArgs] " + errParse.Error())
	}

	_, errAccept := client.AcceptOrderFromCurier(ctx, &pickpoint.AcceptOrderFromCurierRequest{
		Order: domainToOrder(order),
	})
	if errAccept != nil {
		return errors.New("[error in AcceptOrderFromCurier] " + errAccept.Error())
	}

	fmt.Println("\tзаказ с orderId =", order.OrderId, "принят от курьера")

	return nil
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
		Weight:      float64(order.Weight),
		Cost:        order.Package.Cost(),
		Package:     convertPackageType(order.Package),
	}
}

func convertPackageType(pkg models.Package) pickpoint.Order_PackageType {
	switch pkg.Type() {
	case "bag":
		return pickpoint.Order_PACKAGE_TYPE_BAG
	case "box":
		return pickpoint.Order_PACKAGE_TYPE_BOX
	case "film":
		return pickpoint.Order_PACKAGE_TYPE_FILM
	default:
		return pickpoint.Order_PACKAGE_TYPE_UNSPECIFIED
	}
}

func parseAcceptOrderFromCurierArgs(args []string) (models.Order, error) {
	var orderId, clientId int64
	var shelfLife, packageType string
	var weight, cost float64

	fs := flag.NewFlagSet(service.AcceptOrderFromCurier, flag.ContinueOnError)
	fs.Int64Var(&orderId, "orderId", 0, "используйте --orderId=123")
	fs.Int64Var(&clientId, "clientId", 0, "используйте --clientId=123")
	fs.StringVar(&shelfLife, "shelfLife", "", "используйте --shelfLife=2024-05-31T22:42:00+03:00")
	fs.Float64Var(&weight, "weight", 0, "используйте --weight=12.3")
	fs.Float64Var(&cost, "cost", 0, "используйте --cost=12.3")
	fs.StringVar(&packageType, "packageType", "", "используйте --packageType=bag/box/film")
	if err := fs.Parse(args); err != nil {
		return models.Order{}, err
	}

	if orderId <= 0 {
		return models.Order{}, errors.New("orderId должен быть натуральным числом")
	}
	if clientId <= 0 {
		return models.Order{}, errors.New("clientId должен быть натуральным числом")
	}
	if len(shelfLife) == 0 {
		return models.Order{}, errors.New("shelfLife пусто")
	}
	if weight <= 0 {
		return models.Order{}, errors.New("weight должен быть положительным числом")
	}
	if cost <= 0 {
		return models.Order{}, errors.New("cost должен быть положительным числом")
	}
	if len(packageType) == 0 {
		return models.Order{}, errors.New("packageType пусто")
	}

	shelfLifeTime, err := time.Parse(time.RFC3339, shelfLife)
	if err != nil {
		return models.Order{}, err
	}
	if shelfLifeTime.UTC().Before(time.Now().UTC()) {
		return models.Order{}, errors.New("shelfLife в прошлом")
	}

	pkg, err := models.PackageFactory(packageType)
	if err != nil {
		return models.Order{}, err
	}

	order := models.Order{
		OrderId:   models.OrderId(orderId),
		ClientId:  models.ClientId(clientId),
		AddedDate: models.NewAddedDate(time.Now().UTC()),
		ShelfLife: models.NewShelfLife(shelfLifeTime.UTC()),
		Issued:    false,
		Returned:  false,
		Deleted:   false,
		OrderHash: models.Hash(hash.GenerateHash()),
		Weight:    models.Weight(weight),
		Cost:      models.Cost(cost),
		Package:   pkg,
	}

	if err := order.Package.Validate(order); err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func help(ctx context.Context, client pickpoint.PickpointClient) error {
	commands, errHelp := client.Help(ctx, &emptypb.Empty{})
	if errHelp != nil {
		log.Fatal(errHelp)
		return errors.New("[error in Help] " + errHelp.Error())
	}

	fmt.Println("\tСписок команд:")
	for _, cmd := range commands.GetCommands() {
		fmt.Println("\t\t-", cmd.GetName(), cmd.GetDescription())
	}

	return nil
}
