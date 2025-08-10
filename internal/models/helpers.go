package models

import "math"

func RublesToKopecks(rubles float64) int64 {
	return int64(math.Round(rubles * 100))
}

func KopecksToRubles(kopecks int64) float64 {
	return float64(kopecks) / 100
}
