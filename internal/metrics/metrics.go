package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	orderIdLabel     = "orderId"
	pickPointIdLabel = "pickPointId"
)

var (
	numberIssuedOrders = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "number_issued_orders",
		Help: "number of orders that were issued to customers",
	}, []string{
		orderIdLabel,
		pickPointIdLabel,
	})
)

func IncNumberIssuedOrders(orderId, pickPointId string) {
	numberIssuedOrders.With(prometheus.Labels{
		orderIdLabel:     orderId,
		pickPointIdLabel: pickPointId,
	}).Inc()
}
