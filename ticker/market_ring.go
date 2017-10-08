package ticker

import (
	"container/ring"
	"log"
	"net/http"
	"sync"
	"time"
)

const errInvalidMarket = "INVALID_MARKET"

// MarketRing кольцо воркеров
type MarketRing struct {
	*log.Logger
	*ring.Ring
	mu                sync.RWMutex
	maxQueryPerSecond int
	runMu             sync.RWMutex
	runing            bool
}

// MaxQueryPerSecond максимальное количество запросов в секунду
func (mr *MarketRing) MaxQueryPerSecond() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.maxQueryPerSecond
}

// SetMaxQueryPerSecond устанавливает новое значение для максимального количества запросов в секунду
func (mr *MarketRing) SetMaxQueryPerSecond(val int) {
	mr.mu.Lock()
	mr.mu.Unlock()
	mr.maxQueryPerSecond = val
}

// CreateMarketRing создает новое кольцо воркеров
func CreateMarketRing(c *Config, logger *log.Logger) *MarketRing {
	mr := &MarketRing{
		Logger: logger,
		Ring:   ring.New(len(c.Markets)),
	}

	mr.SetMaxQueryPerSecond(c.MaxQueryPerSecond)

	mr.Println("=================== init market ring ====================")
	for i := 0; i < mr.Len(); i++ {
		mr.Println("add worker: ", c.Markets[i])
		mr.Value = CreateWorker(c.Markets[i], mr.Logger)
		mr.Ring = mr.Next()
	}

	return mr
}

func (mr *MarketRing) AddWorker(market string) {
	mr.Stop()
	mr.Println("add worker: ", market)
	newItem := ring.New(1)
	newItem.Value = CreateWorker(market, mr.Logger)
	mr.Link(newItem)
	mr.Run()
}

func (mr *MarketRing) isRuning() bool {
	mr.runMu.RLock()
	defer mr.runMu.RUnlock()

	return mr.runing
}

func (mr *MarketRing) Stop() {
	mr.runMu.Lock()
	defer mr.runMu.Unlock()
	if !mr.runing {
		mr.Println("сервис уже остановлен")
		return
	}

	mr.runing = false
}

func (mr *MarketRing) Run() {
	mr.runMu.Lock()

	if mr.runing {
		mr.Println("служба уже запущена")
		mr.runMu.Unlock()
		return
	}

	mr.runing = true
	mr.runMu.Unlock()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	defer func() {
		client = nil
		mr.Println("все!")
	}()

	t := time.NewTicker(time.Second * 1)

	for _ = range t.C {
		if !mr.isRuning() {
			t.Stop()
			return
		}

		for i := 0; i < mr.MaxQueryPerSecond(); i++ {
			wrk := mr.Value.(*Worker)
			if err := wrk.Exec(client); err != nil {
				mr.Println("'", wrk.market, "' error:", err)

				if err.Error() == errInvalidMarket {
					r := mr.Ring.Unlink(mr.Len() - 1)
					mr.Ring = r
					mr.Println("invalid market '", wrk.market, "' has been deleted")
					wrk = nil
				}
			}

			mr.Ring = mr.Next()
		}
	}
}
