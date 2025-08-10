package models

import "time"

type WithdrawModel struct {
	UserID    string    `json:"-"`
	OrderID   string    `json:"-"`
	Sum       int64     `json:"-"`
	CreatedAt time.Time `json:"-"`
}

func (w *WithdrawModel) ToResponse() *WithdrawResponse {
	return &WithdrawResponse{
		OrderID:     w.OrderID,
		Sum:         KopecksToRubles(w.Sum),
		ProcessedAt: w.CreatedAt,
	}
}

type WithdrawRequest struct {
	OrderID string  `json:"order" binding:"required,numeric"`
	Sum     float64 `json:"sum" binding:"required,gt=0"`
}

func (req *WithdrawRequest) ToModel(userID string) *WithdrawModel {
	return &WithdrawModel{
		UserID:    userID,
		OrderID:   req.OrderID,
		Sum:       RublesToKopecks(req.Sum),
		CreatedAt: time.Now(),
	}
}

type WithdrawResponse struct {
	OrderID     string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type WithdrawModelList []*WithdrawModel

func (list WithdrawModelList) ToResponse() []*WithdrawResponse {
	resp := make([]*WithdrawResponse, len(list))
	for i, item := range list {
		resp[i] = item.ToResponse()
	}
	return resp
}
