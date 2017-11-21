package vmodel

var (
	pricePerFile int
)

// func PriceGet() int {
// 	return CalculatePrice()
// }

func CalculatePrice(ev Event, efs []uint32) int {
	return pricePerFile * len(efs)
}
