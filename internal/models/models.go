package models

import (
	"encoding/json"
	"errors"
	"time"
)

type ClientId int64
type PickPointId int64
type OrderId int64

type Date struct {
	time.Time
}

type AddedDate struct {
	Date
}
type ShelfLife struct {
	Date
}
type IssueDate struct {
	Date
}
type ReturnDate struct {
	Date
}
type DeleteDate struct {
	Date
}

type Limit int

type Page int
type PageSize int

type Hash string

type Weight float64
type Cost float64

type Package interface {
	Type() string
	Cost() float64
	Validate(order Order) error
}

type Order struct {
	OrderId    OrderId    `json:"order_id"`
	ClientId   ClientId   `json:"client_id"`
	AddedDate  AddedDate  `json:"added_date"`
	ShelfLife  ShelfLife  `json:"shelf_life"`
	Issued     bool       `json:"issued"`
	IssueDate  IssueDate  `json:"issue_date"`
	Returned   bool       `json:"returned"`
	ReturnDate ReturnDate `json:"return_date"`
	Deleted    bool       `json:"deleted"`
	DeleteDate DeleteDate `json:"delete_date"`
	OrderHash  Hash       `json:"order_hash"`
	Weight     Weight     `json:"weight"`
	Cost       Cost       `json:"cost"`
	Package    Package    `json:"package"`
}

type PickPoint struct {
	PickPointId PickPointId
	Orders      []Order
}

type OrderWithPickPoint struct {
	Order       Order       `json:"order"`
	PickPointId PickPointId `json:"pick_point_id"`
}

type Data struct {
	PickPoints []PickPoint
}

func NewAddedDate(t time.Time) AddedDate {
	return AddedDate{Date{t}}
}

func NewShelfLife(t time.Time) ShelfLife {
	return ShelfLife{Date{t}}
}

func NewIssueDate(t time.Time) IssueDate {
	return IssueDate{Date{t}}
}

func NewReturnDate(t time.Time) ReturnDate {
	return ReturnDate{Date{t}}
}

func NewDeleteDate(t time.Time) DeleteDate {
	return DeleteDate{Date{t}}
}

func PackageFactory(name string) (Package, error) {
	switch name {
	case "bag":
		return Bag{}, nil
	case "box":
		return Box{}, nil
	case "film":
		return Film{}, nil
	default:
		return nil, errors.New("неизвестный тип упаковки")
	}
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format(time.RFC3339))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	packageType := ""
	if o.Package != nil {
		packageType = o.Package.Type()
	}
	return json.Marshal(&struct {
		PackageType string `json:"package_type"`
		Alias
	}{
		PackageType: packageType,
		Alias:       (Alias)(o),
	})
}

func (o *Order) UnmarshalJSON(data []byte) error {
	type Alias Order
	aux := &struct {
		PackageType string `json:"package_type"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.PackageType != "" {
		pkg, err := PackageFactory(aux.PackageType)
		if err != nil {
			return err
		}
		o.Package = pkg
	} else {
		o.Package = nil
	}
	return nil
}

func (owp OrderWithPickPoint) MarshalJSON() ([]byte, error) {
	type Alias OrderWithPickPoint
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(owp),
	})
}

func (owp *OrderWithPickPoint) UnmarshalJSON(data []byte) error {
	type Alias OrderWithPickPoint
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(owp),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
