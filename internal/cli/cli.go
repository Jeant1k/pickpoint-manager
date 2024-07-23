package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/color"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/hash"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/kafka"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/outbox"
)

type Module interface {
	AcceptOrderFromCurier(ctx context.Context, order models.Order) error
	ReturnOrderToCurier(ctx context.Context, orderId models.OrderId) error
	IssueOrderToClient(ctx context.Context, orderIds map[models.OrderId]bool) error
	ListOrders(ctx context.Context, clientId models.ClientId, limit models.Limit, pickPointId models.PickPointId) ([]models.OrderWithPickPoint, error)
	AcceptReturnFromClient(ctx context.Context, clientId models.ClientId, orderId models.OrderId) error
	ListReturns(ctx context.Context, page models.Page, pageSize models.PageSize) ([]models.Order, error)
}

type Deps struct {
	Module      Module
	PickPointId models.PickPointId
	Outbox      *outbox.Outbox
	Producer    *kafka.Producer
	Consumer    *kafka.Consumer
}

type CLI struct {
	Deps
	commandList []command
}

type Task struct {
	Command string
	Args    []string
	Module  Module
	CLI     CLI
	Ctx     context.Context
}

func (t Task) Execute() {
	switch t.Command {
	case help:
		t.CLI.help(t.Ctx)
	case acceptOrderFromCurier:
		if err := t.CLI.acceptOrderFromCurier(t.Ctx, t.Args); err != nil {
			color.PrintRed("\tError: " + err.Error())
		}
	case returnOrderToCurier:
		if err := t.CLI.returnOrderToCurier(t.Ctx, t.Args); err != nil {
			color.PrintRed("\tError: " + err.Error())
		}
	case issueOrderToClient:
		if err := t.CLI.issueOrderToClient(t.Ctx, t.Args); err != nil {
			color.PrintRed("\tError: " + err.Error())
		}
	case listOrders:
		if err := t.CLI.listOrders(t.Ctx, t.Args); err != nil {
			color.PrintRed("\tError: " + err.Error())
		}
	case acceptReturnFromClient:
		if err := t.CLI.acceptReturnFromClient(t.Ctx, t.Args); err != nil {
			color.PrintRed("\tError: " + err.Error())
		}
	case listReturns:
		if err := t.CLI.listReturns(t.Ctx, t.Args); err != nil {
			color.PrintRed("\tError: " + err.Error())
		}
	default:
		color.PrintRed("\tНеизвестная команда: " + t.Command)
	}
}

// NewCLI создает интерфейс командной строки
func NewCLI(d Deps) CLI {
	cli := CLI{
		Deps: d,
		commandList: []command{
			{
				name:        help,
				description: "вывести справку",
			},
			{
				name:        acceptOrderFromCurier,
				description: "принять заказ от курьера: использование acceptCurier --orderId=123 --clientId=123 --shelfLife=2024-08-20T22:42:00+03:00 --weight=12.3 --cost=12.3 --packageType=bag/box/film",
			},
			{
				name:        returnOrderToCurier,
				description: "вернуть заказ курьеру: использование returnCurier --orderId=123",
			},
			{
				name:        issueOrderToClient,
				description: "выдать заказ клиенту: использование issueClient --orderIds=123,124,125",
			},
			{
				name:        listOrders,
				description: "получить список заказов: использование listOrders --clientId=123 [--limit=10 --pickPointId=123]",
			},
			{
				name:        acceptReturnFromClient,
				description: "принять возврат от клиента: использование acceptClient --clientId=123 --orderId=123",
			},
			{
				name:        listReturns,
				description: "получить список возвратов: использование listReturns --page=1 --pageSize=10",
			},
			{
				name:        exit,
				description: "Закрыть программу",
			},
		},
	}

	return cli
}

// Run запускает интерфейс командной строки
func (c CLI) Run(ctx context.Context, taskQueue chan Task) error {
	if c.Deps.Producer != nil {
		go c.Deps.Outbox.StartBackgroundProcessing(ctx)
		go c.Deps.Consumer.StartConsumer(ctx)
	}

	fmt.Println("\tДля того чтобы узнать что программа умеет, введите help.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%d: ", c.Deps.PickPointId)
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

		if command == exit {
			break
		}

		taskQueue <- Task{
			Command: command,
			Args:    args,
			Module:  c.Deps.Module,
			CLI:     c,
			Ctx:     ctx,
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// acceptOrderFromCurier принимает заказ от курьера. Она на вход принимает ID заказа, ID получателя
// и срок хранения. Заказ нельзя принять дважды. Если срок хранения в прошлом, выдает ошибку. Функция
// если нужно создает файл, записывает в него для нужного PickPointId информацию о заказе.
func (c CLI) acceptOrderFromCurier(ctx context.Context, args []string) error {
	order, errParse := parseAcceptOrderFromCurierArgs(args)
	if errParse != nil {
		return errParse
	}

	errAcceptOrder := c.Module.AcceptOrderFromCurier(ctx, order)
	if errAcceptOrder == nil {
		fmt.Println("\tзаказ с orderId =", order.OrderId, "принят от курьера")

		c.logRequest(ctx, "acceptOrderFromCurier", args)
	}

	return errAcceptOrder
}

func parseAcceptOrderFromCurierArgs(args []string) (models.Order, error) {
	var orderId, clientId int64
	var shelfLife, packageType string
	var weight, cost float64

	fs := flag.NewFlagSet(acceptOrderFromCurier, flag.ContinueOnError)
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

// returnOrderToCurier возвращает заказ курьеру. На вход принимается ID заказа. Функция помечает
// заказ в файле удаленным. Можно вернуть только те заказы, у которых вышел срок хранения и если
// заказы не были выданы клиенту. Также проверяется находится ли заказ в текущем ПВЗ, чтобы нельзя
// было вернуть заказ курьеру из другого ПВЗ.
func (c CLI) returnOrderToCurier(ctx context.Context, args []string) error {
	orderId, errParse := parseReturnOrderToCurierArgs(args)
	if errParse != nil {
		return errParse
	}

	errReturnOrder := c.Module.ReturnOrderToCurier(ctx, orderId)

	if errReturnOrder == nil {
		fmt.Println("\tзаказ с orderId =", orderId, "возвращён курьеру")

		c.logRequest(ctx, "returnOrderToCurier", args)
	}

	return errReturnOrder
}

func parseReturnOrderToCurierArgs(args []string) (models.OrderId, error) {
	var orderId int64

	fs := flag.NewFlagSet(returnOrderToCurier, flag.ContinueOnError)
	fs.Int64Var(&orderId, "orderId", 0, "используйте --orderId=123")
	if err := fs.Parse(args); err != nil {
		return models.OrderId(0), err
	}

	if orderId <= 0 {
		return models.OrderId(0), errors.New("orderId должен быть натуральным числом")
	}

	return models.OrderId(orderId), nil
}

// issueOrderToClient выдает заказ клиенту. На вход принимается список ID заказов. Можно выдавать
// только те заказы, которые были приняты от курьера и чей срок хранения меньше текущей даты. Все ID
// заказов должны принадлежать только одному клиенту. Также проверяется находится ли заказ(ы) в текущем
// ПВЗ, чтобы нельзя было выдать заказ(ы) клиенту из другого ПВЗ.
func (c CLI) issueOrderToClient(ctx context.Context, args []string) error {
	orderIdMap, errParse := parseIssueOrderToClientArgs(args)
	if errParse != nil {
		return errParse
	}

	errIssueOrder := c.Module.IssueOrderToClient(ctx, orderIdMap)

	if errIssueOrder == nil {
		fmt.Print("\tзаказы с orderIds = ")
		for orderId := range orderIdMap {
			fmt.Print(orderId, ", ")
		}
		fmt.Println("выданы клиенту")

		c.logRequest(ctx, "issueOrderToClient", args)
	}

	return errIssueOrder
}

func parseIssueOrderToClientArgs(args []string) (map[models.OrderId]bool, error) {
	var orderIds string

	fs := flag.NewFlagSet(issueOrderToClient, flag.ContinueOnError)
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

// listOrders возаращает список заказов, которые не были выданы, возвращены или удалены. На вход
// принимается ID пользователя как обязательный параметр и опциональные параметры. Параметры позволяют
// получать только последние N заказов или заказы клиента, находящиеся в нашем ПВЗ.
func (c CLI) listOrders(ctx context.Context, args []string) error {
	clientId, limit, pickPointId, errParse := parseListOrdersArgs(args)
	if errParse != nil {
		return errParse
	}

	orders, errListOrders := c.Module.ListOrders(ctx, clientId, limit, pickPointId)
	if errListOrders != nil {
		return errListOrders
	}

	for _, order := range orders {
		fmt.Printf(
			"\tpickPointId: %d, orderId: %d, addedDate: %s, shelfLife: %s\n",
			order.PickPointId,
			order.Order.OrderId,
			order.Order.AddedDate.Local().Format("2006-01-02 15:04:05"),
			order.Order.ShelfLife.Local().Format("2006-01-02 15:04:05"),
		)
	}

	c.logRequest(ctx, "listOrders", args)

	return nil
}

func parseListOrdersArgs(args []string) (models.ClientId, models.Limit, models.PickPointId, error) {
	var clientId, limit, pickPointId int64

	fs := flag.NewFlagSet(listOrders, flag.ContinueOnError)
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

// acceptReturnFromClient принимает возврат от клиента. На вход принимается ID пользователя и ID заказа.
// Заказ может быть возвращен в течение двух дней с момента выдачи. Также она проверяет, что заказ
// выдавался с нашего ПВЗ.
func (c CLI) acceptReturnFromClient(ctx context.Context, args []string) error {
	clientId, orderId, errParse := parseAcceptReturnFromClientArgs(args)
	if errParse != nil {
		return errParse
	}

	errAcceptReturn := c.Module.AcceptReturnFromClient(ctx, models.ClientId(clientId), models.OrderId(orderId))

	if errAcceptReturn == nil {
		fmt.Println("заказ с orderId =", orderId, "возвращен от клиента")

		c.logRequest(ctx, "acceptReturnFromClient", args)
	}

	return errAcceptReturn
}

func parseAcceptReturnFromClientArgs(args []string) (models.ClientId, models.OrderId, error) {
	var clientId, orderId int

	fs := flag.NewFlagSet(acceptReturnFromClient, flag.ContinueOnError)
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

// listReturns возвращает список заказов. Функция выдает список пагинированно.
func (c CLI) listReturns(ctx context.Context, args []string) error {
	page, pageSize, errParse := parseListReturnsArgs(args)
	if errParse != nil {
		return errParse
	}

	returns, errListReturns := c.Module.ListReturns(ctx, models.Page(page), models.PageSize(pageSize))
	if errListReturns != nil {
		return errListReturns
	}

	for _, ret := range returns {
		fmt.Printf(
			"\torderId: %d, clientId: %d, addedDate: %s, returnDate: %s\n",
			ret.OrderId,
			ret.ClientId,
			ret.AddedDate.Local().Format("2006-01-02 15:04:05"),
			ret.ReturnDate.Local().Format("2006-01-02 15:04:05"),
		)
	}

	c.logRequest(ctx, "listReturns", args)

	return nil
}

func parseListReturnsArgs(args []string) (models.Page, models.PageSize, error) {
	var page, pageSize int

	fs := flag.NewFlagSet(listReturns, flag.ContinueOnError)
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

// help выводит справку
func (c CLI) help(ctx context.Context) {
	fmt.Println("\tСписок команд:")
	for _, cmd := range c.commandList {
		fmt.Println("\t\t-", cmd.name, cmd.description)
	}

	c.logRequest(ctx, "help", []string{})
}

func (c CLI) logRequest(ctx context.Context, methodName string, args []string) {
	event := map[string]interface{}{
		"time":  time.Now(),
		"event": methodName,
		"query": args,
	}
	eventJSON, _ := json.Marshal(event)

	if c.Deps.Producer != nil {
		err := c.Deps.Producer.SendMessage("logs", string(eventJSON))
		if err != nil {
			color.PrintRed("\tОшибка отправки события в Kafka: " + err.Error())
			c.Deps.Outbox.AddEvent(ctx, methodName, string(eventJSON))
		}
	} else {
		color.PrintYellow("logs " + string(eventJSON))
	}
}
