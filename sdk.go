package topclient

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-sdks/topclient/signer"
)

// https://open.taobao.com/api.htm?docId=48589&docType=2
// https://open.taobao.com/doc.htm?docId=101617&docType=1

const (
	GatewayURLHTTP  = "http://gw.api.taobao.com/router/rest"
	GatewayURLHTTPS = "https://eco.taobao.com/router/rest"
)

type SDK interface {
	DoRequest(ctx context.Context, req url.Values, resp interface{}) error
}

func New(appKey, appSecret string) SDK {
	return NewEx(GatewayURLHTTP, appKey, appSecret)
}

func NewEx(gatewayURL, appKey, appSecret string) SDK {
	tr := &http.Transport{
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConnsPerHost: 1000,
		TLSHandshakeTimeout: 1 * time.Second,
	}

	return &topSDKImpl{
		gatewayURL: gatewayURL,
		sign:       signer.NewSinger(appKey, appSecret),
		httpCli:    &http.Client{Transport: tr, Timeout: 3 * time.Second},
	}
}

type topSDKImpl struct {
	gatewayURL string
	sign       signer.Signer
	httpCli    *http.Client
}

func (impl *topSDKImpl) DoRequest(ctx context.Context, req url.Values, resp interface{}) (err error) {
	err = impl.sign.Sign(req, signer.SignMethodMD5)
	if err != nil {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", impl.gatewayURL, strings.NewReader(req.Encode()))
	if err != nil {
		return
	}

	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	httpResp, err := impl.httpCli.Do(httpReq)
	if err != nil {
		return
	}

	defer httpResp.Body.Close()

	data, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &resp)

	return
}
