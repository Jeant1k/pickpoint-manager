package models

import "errors"

type Bag struct{}

func (b Bag) Type() string {
	return "bag"
}

func (b Bag) Cost() float64 {
	return 5
}

func (b Bag) Validate(order Order) error {
	if order.Weight >= 10 {
		return errors.New("вес заказа больше или равен 10 кг, пакет не подходит")
	}
	return nil
}
