package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

const (
	baseURL          = "http://free.currencyconverterapi.com/api"
	apiVersion       = "/v5"
	convertEndpoint  = "/convert"
	currency         = "ARS"
	baseCurrency     = "USD"
	currencyValueKey = "val"
	currencyFile     = "currency.txt"
)

var lastCurrency = FetchCurrency(baseCurrency, currency)

func init() {
	if _, err := os.Stat(currencyFile); os.IsNotExist(err) {

	} else {
		// read file for lastCurrency
		fmt.Println("\nreading currency")
		f, err := ioutil.ReadFile(currencyFile)
		if err != nil {
			panic(err)
		}
		lastCurrency, err = strconv.ParseFloat(strings.TrimSpace(string(f)), 64)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	go handleDeath(lastCurrency)
	t := time.NewTicker(30 * time.Minute)
	fmt.Println("Currency:" + strconv.FormatFloat(lastCurrency, 'f', 6, 64))
	for {
		<-t.C
		newCurrency := FetchCurrency(baseCurrency, currency)
		fmt.Println("Currency:" + strconv.FormatFloat(newCurrency, 'f', 6, 64))
		soundAlarm(newCurrency > lastCurrency)
		lastCurrency = newCurrency
	}
}

// Platica structures the response for currency values requests
//{"USD_ARS":{"val":36.6014}}
type Platica struct {
	Rate map[string]float64 `json:"USD_ARS"`
}

// GetRequest from external API
func GetRequest(url string, queryParams string) Platica {
	client := http.Client{}
	completeRequest := url + "?" + queryParams

	request, err := http.NewRequest("GET", completeRequest, nil)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var platicaHoy Platica
	err = json.NewDecoder(resp.Body).Decode(&platicaHoy)
	if err != nil {
		panic(err)
	}
	return platicaHoy
}

// FetchCurrency gets latest currency from the external API
func FetchCurrency(baseCurrency string, currencyString string) float64 {
	queryParams := "q=" + baseCurrency + "_" + currency + "&compact=y"
	platica := GetRequest(baseURL+apiVersion+convertEndpoint, queryParams)
	return platica.Rate[currencyValueKey]
}

func soundAlarm(devaluation bool) error {
	f, err := os.Open("very_nice.mp3")
	if devaluation {
		f, err = os.Open("dollar_alert.mp3")
	}

	if err != nil {
		return err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}
	defer d.Close()

	p, err := oto.NewPlayer(d.SampleRate(), 2, 2, 8192)
	if err != nil {
		return err
	}
	defer p.Close()

	if _, err := io.Copy(p, d); err != nil {
		return err
	}
	return nil
}

func handleDeath(lastCurrency float64) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	sig := <-gracefulStop
	fmt.Printf("\n Caught sig: %+v", sig)
	fmt.Println("\n Writing last update on a piece of paper")
	err := ioutil.WriteFile(currencyFile, []byte(fmt.Sprintf("%s\n", strconv.FormatFloat(lastCurrency, 'f', 6, 64))), 0666)
	if err != nil {
		log.Fatal(err)
	}
	soundAlarm(true)
	os.Exit(0)
}
