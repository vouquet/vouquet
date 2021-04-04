package shop

import (
	"time"
)

type Position interface {
	Id()     string
	Symbol() string
	Size()   float64
	Price()  float64
	OrderType()   string
}

type Fix interface {
	Id()        string
	Symbol()    string
	OrderType() string
	Size()      float64
	Price()     float64
	Yield()     (float64, error)
	Date()      (time.Time, error)
}

type Rate interface {
	Ask()       float64
	Bid()       float64
	High()      float64
	Last()      float64
	Low()       float64
	Symbol()    string
	Time()      time.Time
	Volume()    float64
}
