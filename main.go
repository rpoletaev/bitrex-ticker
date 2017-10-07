package main

import (
	"container/ring"
	"net/http"
	"time"

	"github.com/rpoletaev/bitrex-ticker/ticker"
)

var markets = []string{"BTC-ETH", "BTC-LTC", "BTC-XMR", "BTC-NXT", "BTC-DASH"}

func main() {
	println("=============Init market ring==============")
	marketRing := ring.New(len(markets))
	for i := 0; i < marketRing.Len(); i++ {
		marketRing.Value = ticker.CreateWorker(markets[i], nil)
		marketRing = marketRing.Next()
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	maxQueryPerSecond := 3

	for _ = range time.Tick(time.Second * 1) {
		for i := 0; i < maxQueryPerSecond; i++ {
			marketRing.Value.(*ticker.Worker).Exec(&client)
			marketRing = marketRing.Next()
		}
	}
}
