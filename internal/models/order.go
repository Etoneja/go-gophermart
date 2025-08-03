package models

import (
	"fmt"
	"slices"
	"time"
)

type OrderStatus string
type AccrualOrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

var TerminatedOrderStatuses = []OrderStatus{
	OrderStatusInvalid,
	OrderStatusProcessed,
}

const (
	AccrualOrderStatusRegistered AccrualOrderStatus = "REGISTERED"
	AccrualOrderStatusProcessing AccrualOrderStatus = "PROCESSING"
	AccrualOrderStatusInvalid    AccrualOrderStatus = "INVALID"
	AccrualOrderStatusProcessed  AccrualOrderStatus = "PROCESSED"
)

var AccrualOrderStatusToOrderStatus = map[AccrualOrderStatus]OrderStatus{
	AccrualOrderStatusRegistered: OrderStatusNew,
	AccrualOrderStatusProcessing: OrderStatusInvalid,
	AccrualOrderStatusInvalid:    OrderStatusInvalid,
	AccrualOrderStatusProcessed:  OrderStatusProcessed,
}

func ConvertAccrualOrderStatusToOrderStatus(status AccrualOrderStatus) (OrderStatus, error) {
	if orderStatus, ok := AccrualOrderStatusToOrderStatus[status]; ok {
		return orderStatus, nil
	}
	return "", fmt.Errorf("unknown accrual order status: %v", status)
}

type OrderModel struct {
	ID        string      `json:"-"`
	UserID    string      `json:"-"`
	Status    OrderStatus `json:"-"`
	Accrual   *int64      `json:"-"`
	CreatedAt time.Time   `json:"-"`
	UpdatedAt time.Time   `json:"-"`
}

func (o *OrderModel) IsTerminated() bool {
	return slices.Contains(TerminatedOrderStatuses, o.Status)
}

type OrderResponse struct {
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    *float64    `json:"accrual,omitempty"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

func (o *OrderModel) ToResponse() *OrderResponse {
	resp := &OrderResponse{
		Number:     o.ID,
		Status:     o.Status,
		UploadedAt: o.CreatedAt,
	}
	if o.Accrual != nil {
		val := KopecksToRubles(*o.Accrual)
		resp.Accrual = &val
	}
	return resp
}

type AccrualOrderModel struct {
	ID      string             `json:"-"`
	Status  AccrualOrderStatus `json:"-"`
	Accrual *int64             `json:"-"`
}

type AccrualOrderResponse struct {
	Order   string             `json:"order"`
	Status  AccrualOrderStatus `json:"status"`
	Accrual *float64           `json:"accrual,omitempty"`
}

func (ao *AccrualOrderResponse) ToModel() *AccrualOrderModel {
	model := &AccrualOrderModel{
		ID:     ao.Order,
		Status: ao.Status,
	}
	if ao.Accrual != nil {
		val := RublesToKopecks(*ao.Accrual)
		model.Accrual = &val
	}
	return model
}

type OrderModelList []*OrderModel

func (list OrderModelList) ToResponse() []*OrderResponse {
	resp := make([]*OrderResponse, len(list))
	for i, item := range list {
		resp[i] = item.ToResponse()
	}
	return resp
}
