package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	req, err := http.NewRequest("GET", Binance_Baseurl+"/fapi/v1/depth", nil)
	if err != nil {
		log.Println(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(data))
}
