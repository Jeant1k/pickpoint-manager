package service

const (
	Help                   = "help"         // Вывести справку
	AcceptOrderFromCurier  = "acceptCurier" // Принять заказ от курьера
	ReturnOrderToCurier    = "returnCurier" // Вернуть заказ курьеру
	IssueOrderToClient     = "issueClient"  // Выдать заказ клиенту
	ListOrders             = "listOrders"   // Получить список заказов
	AcceptReturnFromClient = "acceptClient" // Принять возврат от клиента
	ListReturns            = "listReturns"  // Получить список возвратов
	Exit                   = "exit"         // Выход

)

type Command struct {
	Name        string
	Description string
}

// в днях
const ReturnTime = 2
