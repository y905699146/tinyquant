package mod

import (
	"fmt"
	"net/url"
)

type ReqParam struct {
	URL    string
	Method string
	Query  url.Values
	Body   url.Values
	APIKEY string
}

func (r *ReqParam) SetParam(key string, value interface{}) *ReqParam {
	if r.Query == nil {
		r.Query = url.Values{}
	}
	r.Query.Set(key, fmt.Sprintf("%v", value))
	return r
}
