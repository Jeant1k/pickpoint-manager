package models

import "errors"

type Box struct{}

func (b Box) Type() string {
	return "box"
}

func (b Box) Cost() float64 {
	return 20
}

func (b Box) Validate(order Order) error {
	if order.Weight >= 30 {
		return errors.New("вес заказа больше или равен 30 кг, коробка не подходит")
	}
	return nil
}
