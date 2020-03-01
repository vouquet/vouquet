package main

import "github.com/hinoshiba/go-gmo-coin/gomocoin"

func main() {
	API_KEY = "your api key"
	SECRET_KEY = "your secret key"

	gmocoin, err := gomocoin.NewGoMOcoin(API_KEY, SECRET_KEY, context.Background())
	if err != nil {
		panic(err)
	}
	defer gmocoin.Close()

	_, err = gmocoin.UpdateRate()
	if err != nil {
		panic(err)
	}

	//sell
	_, err := gmocoin.Order(gomocoin.SIDE_SELL, gomocoin.SYMBOL_BTC, 0.0001)
	if err != nil {
		panic(err)
	}

	//buy
	_, err = gmocoin.Order(gomocoin.SIDE_BUY, gomocoin.SYMBOL_BTC, 0.0001)
	if err != nil {
		panic(err)
	}
}
