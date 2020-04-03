package quant

type Binance struct {
	Host      string
	ApiKey    string
	SecretKey string
}

func NewBinance(host, apiKey, secretKey string) *Binance {
	return &Binance{
		Host:      host,
		ApiKey:    apiKey,
		SecretKey: secretKey,
	}
}

func GetUserAccount() {

}
