package quant

import (
	"fmt"

	"github.com/huobirdcenter/huobi_golang/pkg/client"
)

func init() {
	// Get the timestamp from Huobi server and print on console
	client := new(client.CommonClient).Init(config.Host)
	resp, err := client.GetTimestamp()

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("timestamp:", resp)
	}

	// Get the list of accounts owned by this API user and print the detail on console
	client := new(client.AccountClient).Init(config.AccessKey, config.SecretKey, config.Host)
	resp, err := client.GetAccountInfo()
	if err != nil {
		fmt.Println(err)
	} else {
		for _, result := range resp {
			fmt.Printf("account: %+v\n", result)
		}
	}
}
