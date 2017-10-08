package ticker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	contentTypeHeader = "Content-Type"
	contentType       = "application/json"
	baseApiURL        = `https://bittrex.com/api/v1.1/public/getticker?market=`
)

// MarketResult структура для результата по тикерам от api
// Bid Ask по условию не нужны, поэтому пусть будут закоменчены
type MarketResult struct {
	// Bid  float64 `json:"Bid"`
	// Ask  float64 `json:"Ask"`
	Last float32 `json:"Last"`
}

type APIResult struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Result  *MarketResult `json:"result"`
}

// Worker хранит информацию по одной паре и осуществляет
// обновление результата по этой паре
type Worker struct {
	*log.Logger
	market    string
	lastValue float32
	url       string
}

// CreateWorker creates new worker and returns worker pointer
func CreateWorker(market string, logger *log.Logger) *Worker {
	if logger == nil {
		logger = log.New(os.Stdout, "", 1)
	}
	return &Worker{
		Logger: logger,
		market: market,
		url:    baseApiURL + market,
	}
}

// Exec запускает выполнение запроса и сравнение результата
func (w *Worker) Exec(client *http.Client) error {
	errCh := make(chan error)

	go func(errChan chan error) {
		defer close(errChan)
		req, err := http.NewRequest(
			http.MethodGet,
			w.url,
			nil,
		)

		if err != nil {
			w.Logger.Println(err)
			errChan <- err
			return
		}

		req.Header.Set(contentTypeHeader, contentType)

		resp, err := client.Do(req)
		if err != nil {
			errChan <- err
			return
		}

		bts, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errChan <- err
			return
		}

		resp.Body.Close()

		var result APIResult
		if err := json.Unmarshal(bts, &result); err != nil {
			errChan <- err
			return
		}

		if !result.Success {
			errChan <- fmt.Errorf("%s", result.Message)
			return
		}

		if w.lastValue != result.Result.Last {
			w.lastValue = result.Result.Last
			w.Logger.Println("'", w.market, "': ", w.lastValue)
		}
	}(errCh)

	return <-errCh
}
