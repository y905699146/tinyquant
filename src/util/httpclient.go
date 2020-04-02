package util

import (
	"encoding/json"
	"fmt"
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
