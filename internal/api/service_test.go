package service

import (
	"testing"
	"time"

	"context"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/models"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/pkg/api/proto/pickpoint/v1/pickpoint/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/api/mocks"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestOrderToDomain(t *testing.T) {
	now := time.Now().UTC()
	req := &pickpoint.AcceptOrderFromCurierRequest{
		Order: &pickpoint.Order{
			OrderId:   &pickpoint.OrderId{OrderId: 1},
			ClientId:  &pickpoint.ClientId{ClientId: 2},
			AddedDate: timestamppb.New(now),
			ShelfLife: timestamppb.New(now.Add(24 * time.Hour)),
			Weight:    9.5,
			Cost:      100.0,
			Package:   pickpoint.Order_PACKAGE_TYPE_BAG,
		},
	}

	expectedOrder := models.Order{
		OrderId:    models.OrderId(1),
		ClientId:   models.ClientId(2),
		AddedDate:  models.NewAddedDate(now),
		ShelfLife:  models.NewShelfLife(now.Add(24 * time.Hour)),
		Weight:     models.Weight(9.5),
		Cost:       models.Cost(100.0),
		Package:    models.Bag{},
		IssueDate:  models.NewIssueDate(time.Unix(0, 0).UTC()),
		ReturnDate: models.NewReturnDate(time.Unix(0, 0).UTC()),
		DeleteDate: models.NewDeleteDate(time.Unix(0, 0).UTC()),
	}

	order, err := orderToDomain(req)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, order)
}

func TestPackageTypeToString(t *testing.T) {
	tests := []struct {
		packageType pickpoint.Order_PackageType
		expected    string
		expectError bool
	}{
		{pickpoint.Order_PACKAGE_TYPE_BAG, "bag", false},
		{pickpoint.Order_PACKAGE_TYPE_BOX, "box", false},
		{pickpoint.Order_PACKAGE_TYPE_FILM, "film", false},
		{pickpoint.Order_PACKAGE_TYPE_UNSPECIFIED, "", true},
	}

	for _, test := range tests {
		result, err := packageTypeToString(&test.packageType)
		if test.expectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
}

func TestPickPointService_RegistratePickPointId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mocks.NewMockModule(ctrl)
	service := &PickPointService{Deps: Deps{Module: mockModule}}

	req := &pickpoint.RegistratePickPointIdRequest{
		PickPointId: &pickpoint.PickPointId{
			PickPointId: 123,
		},
	}
	mockModule.EXPECT().RegistratePickPointId(context.Background(), models.PickPointId(123))

	resp, err := service.RegistratePickPointId(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, &emptypb.Empty{}, resp)
}

func TestPickPointService_AcceptOrderFromCurier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mocks.NewMockModule(ctrl)
	service := &PickPointService{Deps: Deps{Module: mockModule}}

	addedDate := time.Now()

	req := &pickpoint.AcceptOrderFromCurierRequest{
		Order: &pickpoint.Order{
			OrderId:     &pickpoint.OrderId{OrderId: 1},
			ClientId:    &pickpoint.ClientId{ClientId: 10},
			AddedDate:   timestamppb.New(addedDate),
			ShelfLife:   timestamppb.New(addedDate.Add(time.Hour * 10)),
			Issued:      false,
			IssueDate:   &timestamppb.Timestamp{},
			Returned:    false,
			ReturnDate:  &timestamppb.Timestamp{},
			Deleted:     false,
			DeletedDate: &timestamppb.Timestamp{},
			Hash:        "some hash",
			Weight:      3.1415,
			Cost:        100,
			Package:     pickpoint.Order_PACKAGE_TYPE_BOX,
		},
	}

	mockModule.EXPECT().AcceptOrderFromCurier(gomock.Any(), gomock.Any()).Return(nil)

	resp, err := service.AcceptOrderFromCurier(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, &emptypb.Empty{}, resp)
}

func TestPickPointService_ReturnOrderToCurier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mocks.NewMockModule(ctrl)
	service := &PickPointService{Deps: Deps{Module: mockModule}}

	req := &pickpoint.ReturnOrderToCurierRequest{OrderId: &pickpoint.OrderId{OrderId: 1}}

	mockModule.EXPECT().ReturnOrderToCurier(gomock.Any(), gomock.Any()).Return(nil)
	resp, err := service.ReturnOrderToCurier(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, &emptypb.Empty{}, resp)
}

func TestPickPointService_IssueOrderToClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mocks.NewMockModule(ctrl)
	service := &PickPointService{Deps: Deps{Module: mockModule}}

	req := &pickpoint.IssueOrderToClientRequest{
		OrderIds: []*pickpoint.OrderId{
			{OrderId: 1},
			{OrderId: 2},
		},
	}

	mockModule.EXPECT().IssueOrderToClient(gomock.Any(), gomock.Any()).Return(nil)
	resp, err := service.IssueOrderToClient(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, &emptypb.Empty{}, resp)
}

func TestPickPointService_ListOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mocks.NewMockModule(ctrl)
	service := &PickPointService{Deps: Deps{Module: mockModule}}

	addedDate := time.Now()
	orders := []models.OrderWithPickPoint{
		{
			Order: models.Order{
				OrderId:   models.OrderId(1),
				ClientId:  models.ClientId(1),
				AddedDate: models.NewAddedDate(addedDate),
				ShelfLife: models.NewShelfLife(addedDate.Add(time.Hour * 10)),
				Issued:    false,
				Returned:  false,
				Deleted:   false,
				OrderHash: "some hash",
				Weight:    models.Weight(3.1415),
				Cost:      models.Cost(100),
				Package:   models.Box{},
			},
			PickPointId: models.PickPointId(10),
		},
	}

	limit := int64(10)
	req := &pickpoint.ListOrdersRequest{
		ClientId:    &pickpoint.ClientId{ClientId: 1},
		Limit:       &limit,
		PickPointId: &pickpoint.PickPointId{PickPointId: 10},
	}

	mockModule.EXPECT().ListOrders(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(orders, nil)
	resp, err := service.ListOrders(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, len(orders), len(resp.List))
}

func TestPickPointService_AcceptReturnFromClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mocks.NewMockModule(ctrl)
	service := &PickPointService{Deps: Deps{Module: mockModule}}

	req := &pickpoint.AcceptReturnFromClientRequest{
		ClientId: &pickpoint.ClientId{ClientId: 1},
		OrderId:  &pickpoint.OrderId{OrderId: 10},
	}

	mockModule.EXPECT().AcceptReturnFromClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	resp, err := service.AcceptReturnFromClient(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, &emptypb.Empty{}, resp)
}

func TestPickPointService_ListReturns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mocks.NewMockModule(ctrl)
	service := &PickPointService{Deps: Deps{Module: mockModule}}

	addedDate := time.Now()
	returns := []models.Order{
		{
			OrderId:   models.OrderId(1),
			ClientId:  models.ClientId(1),
			AddedDate: models.NewAddedDate(addedDate),
			ShelfLife: models.NewShelfLife(addedDate.Add(time.Hour * 10)),
			Issued:    false,
			Returned:  false,
			Deleted:   false,
			OrderHash: "some hash",
			Weight:    models.Weight(3.1415),
			Cost:      models.Cost(100),
			Package:   models.Box{},
		},
	}

	req := &pickpoint.ListReturnsRequest{
		Page:     1,
		PageSize: 10,
	}

	mockModule.EXPECT().ListReturns(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns, nil)
	resp, err := service.ListReturns(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, len(returns), len(resp.GetOrders()))
}
