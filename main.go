package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"github.com/rpoletaev/bitrex-ticker/ticker"
	"gopkg.in/yaml.v2"
)

var mr *ticker.MarketRing

func main() {
	var configPath string
	flag.StringVar(&configPath, "-c", "config.yaml", "-c=/path/to/config")
	flag.Parse()

	config, err := readConfig(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	l := log.New(os.Stdout, "bitrex-ticker:", 0)
	l.Println(*config)

	mr = ticker.CreateMarketRing(config, l)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/add-market/", addMarketHandler)
	l.Fatal(http.ListenAndServe(":8080", nil))
}

func readConfig(filepath string) (*ticker.Config, error) {
	bts, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	c := &ticker.Config{}
	err = yaml.Unmarshal(bts, c)
	return c, err
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	go mr.Run()
	fmt.Fprintf(w, "server started")
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	go mr.Stop()
	fmt.Fprintf(w, "server stopped")
}

// add new market like /add/USD-XEM
func addMarketHandler(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Path[len("/add-market/"):]
	go mr.AddWorker(market)
	fmt.Fprintf(w, "market '%s' added", market)
}
