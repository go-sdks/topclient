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

type FNDoHTTP func(proc func(client *http.Client))

type DContext struct {
	ReqURL   string
	ReqBody  []byte
	RespBody []byte
	FNDoHTTP FNDoHTTP
}

type SDK interface {
	DoRequest(ctx context.Context, req url.Values, resp interface{}) error
	DoRequestEx(ctx context.Context, req url.Values, resp interface{}, dCtx *DContext) (err error)
}

func New(appKey, appSecret string) SDK {
	return NewEx(GatewayURLHTTP, appKey, appSecret, nil)
}

func NewEx(gatewayURL, appKey, appSecret string, fnDoHTTP FNDoHTTP) SDK {
	tr := &http.Transport{
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConnsPerHost: 1000,
		TLSHandshakeTimeout: 1 * time.Second,
	}

	return &topSDKImpl{
		gatewayURL: gatewayURL,
		sign:       signer.NewSinger(appKey, appSecret),
		httpCli:    &http.Client{Transport: tr, Timeout: 3 * time.Second},
		FNDoHTTP:   fnDoHTTP,
	}
}

type topSDKImpl struct {
	gatewayURL string
	sign       signer.Signer
	httpCli    *http.Client
	FNDoHTTP   FNDoHTTP
}

func (impl *topSDKImpl) DoRequest(ctx context.Context, req url.Values, resp interface{}) (err error) {
	return impl.DoRequestEx(ctx, req, resp, nil)
}

func (impl *topSDKImpl) DoRequestEx(ctx context.Context, req url.Values, resp interface{}, dCtx *DContext) (err error) {
	err = impl.sign.Sign(req, signer.SignMethodMD5)
	if err != nil {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", impl.gatewayURL, strings.NewReader(req.Encode()))
	if err != nil {
		return
	}

	if dCtx != nil {
		dCtx.ReqURL = impl.gatewayURL
		dCtx.ReqBody = []byte(req.Encode())
	}

	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	fnDo := func(client *http.Client) (data []byte, err error) {
		httpResp, err := client.Do(httpReq)
		if err != nil {
			return
		}
		defer httpResp.Body.Close()

		return ioutil.ReadAll(httpResp.Body)
	}

	var data []byte

	if impl.FNDoHTTP != nil {
		impl.FNDoHTTP(func(client *http.Client) {
			data, err = fnDo(client)
		})
	} else {
		data, err = fnDo(impl.httpCli)
	}

	if err != nil {
		return
	}

	if dCtx != nil {
		dCtx.RespBody = data
	}

	err = json.Unmarshal(data, &resp)

	return
}
