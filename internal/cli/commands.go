package cli

const (
	help                   = "help"         // Вывести справку
	acceptOrderFromCurier  = "acceptCurier" // Принять заказ от курьера
	returnOrderToCurier    = "returnCurier" // Вернуть заказ курьеру
	issueOrderToClient     = "issueClient"  // Выдать заказ клиенту
	listOrders             = "listOrders"   // Получить список заказов
	acceptReturnFromClient = "acceptClient" // Принять возврат от клиента
	listReturns            = "listReturns"  // Получить список возвратов
	exit                   = "exit"         // Выход

)

type command struct {
	name        string
	description string
}

// в днях
const ReturnTime = 2
