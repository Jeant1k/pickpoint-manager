package tests

import (
	"time"

	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
)

var (
	order1 = models.Order{
		OrderId:   models.OrderId(123),
		ClientId:  models.ClientId(123),
		AddedDate: models.NewAddedDate(time.Now().UTC()),
		ShelfLife: models.NewShelfLife(time.Now().Add(24 * time.Hour).UTC()),
		Issued:    false,
		Returned:  false,
		Deleted:   false,
		OrderHash: "somehash",
		Weight:    models.Weight(5.55),
		Cost:      models.Cost(10),
		Package:   models.Box{},
	}
	pickPointId1 = models.PickPointId(1)
)
