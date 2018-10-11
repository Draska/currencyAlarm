package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

const baseURL = "http://data.fixer.io/api/"
const latestEndpoint = "latest"
const convertEndpoint = "convert"
const currency = "ARS"
const baseCurrency = "USD"
const amountToConvert = "1"
const accessKey = "c1685baafe05822358c74bad1c35c550"

func main() {
	t := time.NewTicker(time.Hour)
	for {
		yesterday := time.Now().Local().AddDate(0, 0, -1)
		yesterString := yesterday.Format("2006-01-02")
		requestToday := baseURL + latestEndpoint + "?access_key=" + accessKey + "&symbols=" + currency
		requestYesterday := baseURL + yesterString + "?access_key=" + accessKey + "&symbols=" + currency
		MakeRequest(requestToday, requestYesterday)
		<-t.C
	}
}

// Platica structures the response for currency values requests
type Platica struct {
	Success   bool               `json:"success"`
	Timestamp int                `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

//{"success":true,"timestamp":1539268446,"base":"EUR","date":"2018-10-11","rates":{"ARS":42.65679}}

// MakeRequest asks for stuff
func MakeRequest(requestToday string, requestYesterday string) {
	client := http.Client{}

	request, err := http.NewRequest("GET", requestToday, nil)
	if err != nil {
		log.Fatalln(err)
	}

	request2, err := http.NewRequest("GET", requestYesterday, nil)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	respYesterday, err2 := client.Do(request2)
	if err2 != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	defer respYesterday.Body.Close()

	var platicaHoy Platica
	var platicaAyer Platica

	err = json.NewDecoder(resp.Body).Decode(&platicaHoy)
	if err != nil {
		panic(err)
	}

	err = json.NewDecoder(respYesterday.Body).Decode(&platicaAyer)
	if err != nil {
		panic(err)
	}
	if platicaHoy.Rates["ARS"] >= platicaAyer.Rates["ARS"] {
		soundAlarm()
	} else if platicaHoy.Rates["ARS"] < platicaAyer.Rates["ARS"] {
		soundGoodAlarm()
	}
	log.Println("TODAY")
	log.Println(platicaHoy.Rates["ARS"])
	log.Println("YESTERDAY")
	log.Println(platicaAyer.Rates["ARS"])
}

func soundAlarm() error {
	f, err := os.Open("dollar_alert.mp3")
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

	fmt.Printf("Length: %d[bytes]\n", d.Length())

	if _, err := io.Copy(p, d); err != nil {
		return err
	}
	return nil
}

func soundGoodAlarm() error {
	f, err := os.Open("very_nice.mp3")
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

	fmt.Printf("Length: %d[bytes]\n", d.Length())

	if _, err := io.Copy(p, d); err != nil {
		return err
	}
	return nil
}
