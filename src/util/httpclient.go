package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func InitHttpClient(url string, data interface{}) {
	d, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	body := strings.NewReader(string(d))
	req, err := http.NewRequest("POST", url, body)
	clt := http.Client{}
	clt.Do(req)
}

func HttpGet() {

	url := "https://api.binance.com/api/v3/time"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(body))
}
