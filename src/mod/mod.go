package mod

import (
	"fmt"
	"net/http"
	"net/url"
)

type ReqParam struct {
	URL    string
	Method string
	Query  url.Values
	Body   url.Values
	Header http.Header
}

func (r *ReqParam) SetParam(key string, value interface{}) *ReqParam {
	if r.Query == nil {
		r.Query = url.Values{}
	}
	r.Query.Set(key, fmt.Sprintf("%v", value))
	return r
}
