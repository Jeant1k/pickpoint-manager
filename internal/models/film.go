package models

type Film struct{}

func (f Film) Type() string {
	return "film"
}

func (f Film) Cost() float64 {
	return 1
}

func (f Film) Validate(order Order) error {
	return nil
}
