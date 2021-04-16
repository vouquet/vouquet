package shop

import (
	"math/big"
)

const (
	BIT_SIZE uint = 128
	SATOSHI_MUL int = 1000000000
)

func float64Add(x float64, y float64) float64 {
	return toBtc(toSatoshi(x) + toSatoshi(y))
}

func float64Sub(x float64, y float64) float64 {
	return toBtc(toSatoshi(x) - toSatoshi(y))
}

func float64Mul(x float64, y float64) float64 {
	return toBtc(toSatoshi(x) * toSatoshi(y))
}

func float64Div(x float64, y float64) float64 {
	return toBtc(toSatoshi(x) / toSatoshi(y))
}

func toSatoshi(x float64) int64 {
	b_x, b_y := fl64tBig(x, float64(SATOSHI_MUL))
	return int64(big2Fl64(new(big.Float).SetPrec(BIT_SIZE).Mul(b_x, b_y)))
}

func toBtc(x int64) float64 {
	b_x, b_y := fl64tBig(float64(x), float64(SATOSHI_MUL))
	return big2Fl64(new(big.Float).SetPrec(BIT_SIZE).Quo(b_x, b_y))
}

func fl64tBig(x, y float64) (*big.Float, *big.Float) {
	return big.NewFloat(x), big.NewFloat(y)
}

func big2Fl64(x *big.Float) float64 {
	f64_x, _ := x.Float64()
	return f64_x
}
