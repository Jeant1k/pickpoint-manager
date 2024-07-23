package cli

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
)

func TestParseAcceptOrderFromCurierArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr error
	}{
		{
			name:        "missing orderId",
			args:        []string{"--clientId=1", "--shelfLife=2024-07-20T22:42:00+03:00", "--weight=12.3", "--cost=12.3", "--packageType=box"},
			expectedErr: errors.New("orderId должен быть натуральным числом"),
		},
		{
			name:        "missing clientId",
			args:        []string{"--orderId=1", "--shelfLife=2024-07-20T22:42:00+03:00", "--weight=12.3", "--cost=12.3", "--packageType=box"},
			expectedErr: errors.New("clientId должен быть натуральным числом"),
		},
		{
			name:        "missing shelfLife",
			args:        []string{"--orderId=1", "--clientId=1", "--weight=12.3", "--cost=12.3", "--packageType=box"},
			expectedErr: errors.New("shelfLife пусто"),
		},
		{
			name:        "missing weight",
			args:        []string{"--orderId=1", "--clientId=1", "--shelfLife=2024-07-20T22:42:00+03:00", "--cost=12.3", "--packageType=box"},
			expectedErr: errors.New("weight должен быть положительным числом"),
		},
		{
			name:        "missing cost",
			args:        []string{"--orderId=1", "--clientId=1", "--shelfLife=2024-07-20T22:42:00+03:00", "--weight=12.3", "--packageType=box"},
			expectedErr: errors.New("cost должен быть положительным числом"),
		},
		{
			name:        "missing packageType",
			args:        []string{"--orderId=1", "--clientId=1", "--shelfLife=2024-07-20T22:42:00+03:00", "--weight=12.3", "--cost=12.3"},
			expectedErr: errors.New("packageType пусто"),
		},
		{
			name:        "invalid shelfLife format",
			args:        []string{"--orderId=1", "--clientId=1", "--shelfLife=invalid-date", "--weight=12.3", "--cost=12.3", "--packageType=box"},
			expectedErr: errors.New("parsing time \"invalid-date\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid-date\" as \"2006\""),
		},
		{
			name:        "shelfLife in the past",
			args:        []string{"--orderId=1", "--clientId=1", "--shelfLife=2022-07-20T22:42:00+03:00", "--weight=12.3", "--cost=12.3", "--packageType=box"},
			expectedErr: errors.New("shelfLife в прошлом"),
		},
		{
			name: "valid arguments",
			args: []string{"--orderId=123000", "--clientId=1", "--shelfLife=2024-07-20T22:42:00+03:00", "--weight=12.3", "--cost=120000.3", "--packageType=box"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			order, err := parseAcceptOrderFromCurierArgs(tt.args)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, models.OrderId(123000), order.OrderId)
				assert.Equal(t, models.ClientId(1), order.ClientId)
				assert.Equal(t, models.Weight(12.3), order.Weight)
				assert.Equal(t, models.Cost(120000.3), order.Cost)
				assert.Equal(t, "box", order.Package.Type())
				assert.WithinDuration(t, time.Date(2024, 7, 20, 19, 42, 0, 0, time.UTC), order.ShelfLife.Time, time.Second)
				assert.Equal(t, false, order.Issued)
				assert.Equal(t, false, order.Returned)
				assert.Equal(t, false, order.Deleted)
				assert.Equal(t, models.Box{}, order.Package)

			}
		})
	}
}

func TestParseReturnOrderToCurierArgsArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr error
	}{
		{
			name:        "missing orderId",
			args:        []string{},
			expectedErr: errors.New("orderId должен быть натуральным числом"),
		},
		{
			name: "valid arguments",
			args: []string{"--orderId=123"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			orderId, err := parseReturnOrderToCurierArgs(tt.args)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, models.OrderId(123), orderId)
			}
		})
	}
}

func TestParseIssueOrderToClientArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr error
	}{
		{
			name:        "missing orderIds",
			args:        []string{"--orderIds="},
			expectedErr: errors.New("orderIds пусто"),
		},
		{
			name:        "non natural orderId",
			args:        []string{"--orderIds=123,-124"},
			expectedErr: errors.New("orderId должен быть натуральным числом"),
		},
		{
			name:        "duplication orderId",
			args:        []string{"--orderIds=123,123,124"},
			expectedErr: errors.New("дублирование orderId"),
		},
		{
			name: "valid arguments",
			args: []string{"--orderIds=123,124,125"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			orderIdMap, err := parseIssueOrderToClientArgs(tt.args)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, map[models.OrderId]bool{
					models.OrderId(123): true,
					models.OrderId(124): true,
					models.OrderId(125): true,
				}, orderIdMap)
			}
		})
	}
}

func TestParseListOrdersArgs(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		expectedErr         error
		expectedLimit       models.Limit
		expectedPickPointId models.PickPointId
	}{
		{
			name:        "missing clientId",
			args:        []string{"--limit=10", "--pickPointId=123"},
			expectedErr: errors.New("clientId должен быть натуральным числом"),
		},
		{
			name:        "negative limit",
			args:        []string{"--clientId=123", "--limit=-10", "--pickPointId=123"},
			expectedErr: errors.New("limit должен быть натуральным числом"),
		},
		{
			name:        "negative pickPointId",
			args:        []string{"--clientId=123", "--limit=10", "--pickPointId=-123"},
			expectedErr: errors.New("pickPointId должен быть натуральным числом"),
		},
		{
			name:                "valid arguments (without optional)",
			args:                []string{"--clientId=123000"},
			expectedLimit:       models.Limit(0),
			expectedPickPointId: models.PickPointId(0),
		},
		{
			name:                "valid arguments (with optional)",
			args:                []string{"--clientId=123000", "--limit=10", "--pickPointId=123"},
			expectedLimit:       models.Limit(10),
			expectedPickPointId: models.PickPointId(123),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientId, limit, pickPointId, err := parseListOrdersArgs(tt.args)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, models.ClientId(123000), clientId)
				assert.Equal(t, tt.expectedLimit, limit)
				assert.Equal(t, tt.expectedPickPointId, pickPointId)
			}
		})
	}
}

func TestParseAcceptReturnFromClientArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr error
	}{
		{
			name:        "missing clientId",
			args:        []string{"--orderId=123"},
			expectedErr: errors.New("clientId должен быть натуральным числом"),
		},
		{
			name:        "missing orderId",
			args:        []string{"--clientId=123"},
			expectedErr: errors.New("orderId должен быть натуральным числом"),
		},
		{
			name: "valid arguments",
			args: []string{"--clientId=123000", "--orderId=123"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientId, orderId, err := parseAcceptReturnFromClientArgs(tt.args)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, models.ClientId(123000), clientId)
				assert.Equal(t, models.OrderId(123), orderId)
			}
		})
	}
}

func TestParseListReturnsArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr error
	}{
		{
			name:        "negative pageSize",
			args:        []string{"--pageSize=-10", "--page=1"},
			expectedErr: errors.New("pageSize должен быть натуральным числом"),
		},
		{
			name:        "negative page",
			args:        []string{"--pageSize=10", "--page=-1"},
			expectedErr: errors.New("page должен быть натуральным числом"),
		},
		{
			name: "valid arguments",
			args: []string{"--page=1", "--pageSize=10"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			page, pageSize, err := parseListReturnsArgs(tt.args)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, models.Page(1), page)
				assert.Equal(t, models.PageSize(10), pageSize)
			}
		})
	}
}
