package ticker

import (
	"encoding/json"
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

type Worker struct {
	*log.Logger
	market    string
	lastValue float32
	url       string
}

// CreateWorker creates new worker and returns worker pointer
func CreateWorker(market string, logger *log.Logger) *Worker {
	if logger == nil {
		logger = log.New(os.Stdout, market+"_worker", 1)
	}
	return &Worker{
		Logger: logger,
		market: market,
		url:    baseApiURL + market,
	}
}

func (w *Worker) Exec(client *http.Client) {
	go func() {
		req, err := http.NewRequest(
			http.MethodGet,
			w.url,
			nil,
		)

		if err != nil {
			w.Logger.Println(err)
			return
		}

		req.Header.Set(contentTypeHeader, contentType)

		resp, err := client.Do(req)
		if err != nil {
			w.Logger.Println(err)
			return
		}

		bts, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.Logger.Println(err)
			return
		}

		resp.Body.Close()

		var result APIResult
		if err := json.Unmarshal(bts, &result); err != nil {
			w.Logger.Println(err)
			return
		}

		if !result.Success {
			w.Logger.Printf("market '%s' error: %s\n", w.market, result.Message)
			return
		}

		if w.lastValue != result.Result.Last {
			w.lastValue = result.Result.Last
			w.Logger.Printf("'%s': %d\n", w.market, w.lastValue)
		}
	}()
	// fmt.Printf("%+v\n", result)
}
