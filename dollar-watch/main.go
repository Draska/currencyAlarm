package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

const baseURL = "http://free.currencyconverterapi.com/api"
const apiVersion = "/v5"
const convertEndpoint = "/convert"
const currency = "ARS"
const baseCurrency = "USD"
const currencyValueKey = "val"

func main() {
	lastCurrency := FetchCurrency(baseCurrency, currency)
	t := time.NewTicker(30 * time.Minute)
	log.Println("Currency:" + strconv.FormatFloat(lastCurrency, 'f', 6, 64))
	for {
		<-t.C
		newCurrency := FetchCurrency(baseCurrency, currency)
		log.Println("Currency:" + strconv.FormatFloat(newCurrency, 'f', 6, 64))
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
