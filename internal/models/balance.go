package models

type BalanceModel struct {
	Current   int64 `json:"-"`
	Withdrawn int64 `json:"-"`
}

func (b *BalanceModel) ToResponse() BalanceResponse {
	return BalanceResponse{
		Current:   KopecksToRubles(b.Current),
		Withdrawn: KopecksToRubles(b.Withdrawn),
	}
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
