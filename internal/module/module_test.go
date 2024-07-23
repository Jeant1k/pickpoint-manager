package module

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/cli"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
)

func TestValidateAcceptOrderFromCurier(t *testing.T) {
	tests := []struct {
		name        string
		order       models.Order
		expectedErr error
	}{
		{
			name:        "order already exists",
			order:       models.Order{OrderId: 1},
			expectedErr: errors.New("заказ уже существует"),
		},
		{
			name:        "order does not exist",
			order:       models.Order{OrderId: 0},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateAcceptOrderFromCurier(tt.order)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateReturnOrderToCurier(t *testing.T) {
	tests := []struct {
		name        string
		order       models.Order
		expectedErr error
	}{
		{
			name:        "order not found",
			order:       models.Order{OrderId: 0},
			expectedErr: errors.New("заказ не найден"),
		},
		{
			name:        "order issued to client",
			order:       models.Order{OrderId: 1, Issued: true},
			expectedErr: errors.New("заказ был выдан клиенту"),
		},
		{
			name:        "shelfLife in future",
			order:       models.Order{OrderId: 1, ShelfLife: models.NewShelfLife(time.Now().Add(24 * time.Hour))},
			expectedErr: errors.New("shelfLife в будущем"),
		},
		{
			name:        "order deleted",
			order:       models.Order{OrderId: 1, Deleted: true},
			expectedErr: errors.New("заказ был возвращен курьеру"),
		},
		{
			name: "valid order",
			order: models.Order{
				OrderId:   1,
				Issued:    false,
				ShelfLife: models.NewShelfLife(time.Now().Add(-24 * time.Hour)),
				Deleted:   false,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateReturnOrderToCurier(tt.order)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateIssueOrderToClient(t *testing.T) {
	tests := []struct {
		name        string
		order       models.Order
		expectedErr error
	}{
		{
			name:        "order not found",
			order:       models.Order{OrderId: 0},
			expectedErr: errors.New("заказ не найден"),
		},
		{
			name:        "order already issued",
			order:       models.Order{OrderId: 1, Issued: true},
			expectedErr: errors.New("заказ был выдан клиенту"),
		},
		{
			name:        "order returned",
			order:       models.Order{OrderId: 1, Returned: true},
			expectedErr: errors.New("заказ был возвращен клиентом"),
		},
		{
			name:        "order deleted",
			order:       models.Order{OrderId: 1, Deleted: true},
			expectedErr: errors.New("заказ был возвращен курьеру"),
		},
		{
			name:        "shelfLife in past",
			order:       models.Order{OrderId: 1, ShelfLife: models.NewShelfLife(time.Now().Add(-24 * time.Hour))},
			expectedErr: errors.New("shelfLife в прошлом"),
		},
		{
			name: "valid order",
			order: models.Order{
				OrderId:   1,
				Issued:    false,
				Returned:  false,
				ShelfLife: models.NewShelfLife(time.Now().Add(24 * time.Hour)),
				Deleted:   false,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateIssueOrderToClient(tt.order)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAcceptReturnFromClient(t *testing.T) {
	tests := []struct {
		name        string
		order       models.Order
		expectedErr error
	}{
		{
			name:        "order not found",
			order:       models.Order{OrderId: 0},
			expectedErr: errors.New("заказ не найден"),
		},
		{
			name:        "order not issued",
			order:       models.Order{OrderId: 1, Issued: false},
			expectedErr: errors.New("заказ не был выдан клиенту"),
		},
		{
			name:        "order returned",
			order:       models.Order{OrderId: 1, Issued: true, Returned: true},
			expectedErr: errors.New("заказ был возвращен клиентом"),
		},
		{
			name:        "order deleted",
			order:       models.Order{OrderId: 1, Issued: true, Returned: false, Deleted: true},
			expectedErr: errors.New("заказ был возвращен курьеру"),
		},
		{
			name:        "return time expired",
			order:       models.Order{OrderId: 1, Issued: true, Returned: false, Deleted: false, IssueDate: models.NewIssueDate(time.Now().Add(-cli.ReturnTime*24*time.Hour - time.Hour))},
			expectedErr: errors.New("время возврата просрочено"),
		},
		{
			name: "valid return order",
			order: models.Order{
				OrderId:   1,
				Issued:    true,
				Returned:  false,
				Deleted:   false,
				IssueDate: models.NewIssueDate(time.Now().Add(-cli.ReturnTime*24*time.Hour + time.Hour)),
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateAcceptReturnFromClient(tt.order)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
