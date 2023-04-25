package httpClient

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/go-opener/ctxflow/layer"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const (
	HttpHeaderService = "SERVICE"
	// trace 日志前缀标识（放在[]zap.Field的第一个位置提高效率）
	TopicType = "_tp"
	// 业务日志名字
	LogNameServer = "server"
	// access 日志文件名字
	LogNameAccess = "access"
	// module 日志文件名字
	LogNameModule = "module"

	TraceHeaderKey      = "Uber-Trace-Id"
	LogIDHeaderKey      = "X_BD_LOGID"
	LogIDHeaderKeyLower = "x_bd_logid"
)

var globalTransport *http.Transport

func init() {
	globalTransport = &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     300 * time.Second,
	}
}

type ApiConf struct {
	Service        string        `yaml:"service"`
	AppKey         string        `yaml:"appkey"`
	Domain         string        `yaml:"domain"`
	Timeout        time.Duration `yaml:"timeout"`
	ConnectTimeout time.Duration `yaml:"connectTimeout"`
	Retry          int           `yaml:"retry"`
	HttpStat       bool          `yaml:"httpStat"`
	Proxy          string        `yaml:"proxy"`
	BasicAuth      struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}
}

func (entity *ApiConf) GetTransPort() *http.Transport {
	trans := globalTransport
	if entity.Proxy != "" {
		trans.Proxy = func(_ *http.Request) (*url.URL, error) {
			return url.Parse(entity.Proxy)
		}
	} else {
		trans.Proxy = nil
	}

	if entity.ConnectTimeout != 0 {
		trans.DialContext = (&net.Dialer{
			Timeout: entity.ConnectTimeout,
		}).DialContext
	} else {
		trans.DialContext = nil
	}

	return trans
}

type HttpClient struct {
	layer.Flow
	ApiConf ApiConf
}

func (entity *HttpClient) PreUse(args ...interface{}) {
	entity.Flow.PreUse(args...)
}

func (entity *HttpClient) MakeRequest(method, url string, data io.Reader, headers map[string]string, cookies map[string]string, bodyType string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if h := req.Header.Get("host"); h != "" {
		req.Host = h
	}

	for k, v := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	if entity.ApiConf.BasicAuth.Username != "" {
		req.SetBasicAuth(entity.ApiConf.BasicAuth.Username, entity.ApiConf.BasicAuth.Password)
	}

	cType := bodyType
	if cType == "" { // 根据 encode 获得一个默认的类型
		cType = "application/x-www-form-urlencoded"
	}
	req.Header.Set("Content-Type", cType)

	req.Header.Set(HttpHeaderService, entity.GetLogCtx().AppName)
	req.Header.Set(TraceHeaderKey, entity.GetLogCtx().ReqId)

	req.Header.Set(LogIDHeaderKey, entity.GetLogCtx().LogId)
	req.Header.Set(LogIDHeaderKeyLower, entity.GetLogCtx().LogId)

	return req, nil
}

type timeTrace struct {
	dnsStartTime,
	dnsDoneTime,
	connectDoneTime,
	gotConnTime,
	gotFirstRespTime,
	tlsHandshakeStartTime,
	tlsHandshakeDoneTime,
	finishTime time.Time
}

func (entity *HttpClient) beforeHttpStat(req *http.Request) *timeTrace {
	if entity.ApiConf.HttpStat == false {
		return nil
	}

	var t = &timeTrace{}
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t.dnsStartTime = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t.dnsDoneTime = time.Now() },
		ConnectStart: func(_, _ string) {
			if t.dnsDoneTime.IsZero() {
				t.dnsDoneTime = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			t.connectDoneTime = time.Now()
		},
		GotConn:              func(_ httptrace.GotConnInfo) { t.gotConnTime = time.Now() },
		GotFirstResponseByte: func() { t.gotFirstRespTime = time.Now() },
		TLSHandshakeStart:    func() { t.tlsHandshakeStartTime = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t.tlsHandshakeDoneTime = time.Now() },
	}
	*req = *req.WithContext(httptrace.WithClientTrace(context.Background(), trace))
	return t
}

func (entity *HttpClient) afterHttpStat(scheme string, t *timeTrace) {
	if entity.ApiConf.HttpStat == false {
		return
	}
	t.finishTime = time.Now() // after read body

	if t.dnsStartTime.IsZero() {
		t.dnsStartTime = t.dnsDoneTime
	}

	cost := func(d time.Duration) float64 {
		if d < 0 {
			return -1
		}
		return float64(d.Nanoseconds()/1e4) / 100.0
	}

	switch scheme {
	case "https":
		f := []zap.Field{
			zap.Float64("dnsLookupCost", cost(t.dnsDoneTime.Sub(t.dnsStartTime))),                       // dns lookup
			zap.Float64("tcpConnectCost", cost(t.connectDoneTime.Sub(t.dnsDoneTime))),                   // tcp connection
			zap.Float64("tlsHandshakeCost", cost(t.tlsHandshakeStartTime.Sub(t.tlsHandshakeStartTime))), // tls handshake
			zap.Float64("serverProcessCost", cost(t.gotFirstRespTime.Sub(t.gotConnTime))),               // server processing
			zap.Float64("contentTransferCost", cost(t.finishTime.Sub(t.gotFirstRespTime))),              // content transfer
			zap.Float64("totalCost", cost(t.finishTime.Sub(t.dnsStartTime))),                            // total cost
		}
		entity.GetLog().Desugar().Warn("time trace", f...)
	case "http":
		f := []zap.Field{
			zap.Float64("dnsLookupCost", cost(t.dnsDoneTime.Sub(t.dnsStartTime))),          // dns lookup
			zap.Float64("tcpConnectCost", cost(t.gotConnTime.Sub(t.dnsDoneTime))),          // tcp connection
			zap.Float64("serverProcessCost", cost(t.gotFirstRespTime.Sub(t.gotConnTime))),  // server processing
			zap.Float64("contentTransferCost", cost(t.finishTime.Sub(t.gotFirstRespTime))), // content transfer
			zap.Float64("totalCost", cost(t.finishTime.Sub(t.dnsStartTime))),               // total cost
		}
		entity.GetLog().Desugar().Warn("time trace", f...)
	}
}

func GetFormatRequestTime(time time.Time) string {
	return fmt.Sprintf("%d.%d", time.Unix(), time.Nanosecond()/1e3)
}

func GetRequestCost(start, end time.Time) float64 {
	return float64(end.Sub(start).Nanoseconds()/1e4) / 100.0
}

// 本次请求正确性判断
func calRalCode(resp *http.Response, err error) int {
	if err != nil || resp == nil || resp.StatusCode >= 400 || resp.StatusCode == 0 {
		return -1
	}
	return 0
}

func (entity *HttpClient) httpDo(apiReq *ApiRequest) (httpCode int, response []byte, field []zap.Field, err error) {
	start := time.Now()
	fields := []zap.Field{
		zap.String(TopicType, LogNameModule),
		zap.String("prot", "http"),
		zap.String("service", entity.ApiConf.Service),
		zap.String("method", apiReq.Req.Method),
		zap.String("domain", entity.ApiConf.Domain),
		zap.String("requestUri", apiReq.Req.URL.Path),
		zap.String("proxy", entity.ApiConf.Proxy),
		zap.Duration("timeout", entity.ApiConf.Timeout),
		zap.String("requestStartTime", GetFormatRequestTime(start)),
	}

	apiReq.clientInit.Do(func() {
		if apiReq.HTTPClient == nil {
			timeout := 3 * time.Second
			if entity.ApiConf.Timeout > 0 {
				timeout = entity.ApiConf.Timeout
			}

			trans := entity.ApiConf.GetTransPort()
			apiReq.HTTPClient = &http.Client{
				Timeout:   timeout,
				Transport: trans,
			}
		}
	})

	var (
		resp         *http.Response
		dataBuffer   *bytes.Reader
		maxAttempts  int
		attemptCount int
		doErr        error
		shouldRetry  bool
	)

	attemptCount, maxAttempts = 0, entity.ApiConf.Retry

	retryPolicy := apiReq.GetRetryPolicy()
	backOffPolicy := apiReq.GetBackOffPolicy()

	for {
		if apiReq.Req.GetBody != nil {
			bodyReadCloser, _ := apiReq.Req.GetBody()
			apiReq.Req.Body = bodyReadCloser
		} else if apiReq.Req.Body != nil {
			if dataBuffer == nil {
				data, err := ioutil.ReadAll(apiReq.Req.Body)
				_ = apiReq.Req.Body.Close()
				if err != nil {
					return 0, []byte{}, fields, err
				}
				dataBuffer = bytes.NewReader(data)
				apiReq.Req.ContentLength = int64(dataBuffer.Len())
				apiReq.Req.Body = ioutil.NopCloser(dataBuffer)
			}
			_, _ = dataBuffer.Seek(0, io.SeekStart)
		}

		attemptCount++
		resp, doErr = apiReq.HTTPClient.Do(apiReq.Req)
		if doErr != nil {
			f := []zap.Field{
				zap.String("prot", "http"),
				zap.String("service", entity.ApiConf.Service),
				zap.String("requestUri", apiReq.Req.URL.Path),
				zap.Duration("timeout", entity.ApiConf.Timeout),
				zap.Int("attemptCount", attemptCount),
			}
			entity.GetLog().Desugar().Warn(doErr.Error(), f...)
		}

		shouldRetry = retryPolicy(resp, doErr)
		if !shouldRetry {
			break
		}

		if attemptCount > maxAttempts {
			break
		}

		if doErr == nil {
			drainAndCloseBody(resp, 16384)
		}
		wait := backOffPolicy(attemptCount)
		select {
		case <-apiReq.Req.Context().Done():
			return 0, []byte{}, fields, apiReq.Req.Context().Err()
		case <-time.After(wait):
		}
	}

	if resp != nil {
		httpCode = resp.StatusCode
		response, err = ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}

	err = doErr
	if err == nil && shouldRetry {
		err = fmt.Errorf("hit retry policy")
	}

	end := time.Now()
	if err != nil {
		err = fmt.Errorf("giving up after %d attempt(s): %w", attemptCount, err)
	}

	fields = append(fields,
		//zap.String("retry", fmt.Sprintf("%d/%d", attemptCount-1, client.Retry)),
		zap.Int("httpCode", httpCode),
		zap.String("requestEndTime", GetFormatRequestTime(end)),
		zap.Float64("cost", GetRequestCost(start, end)),
		zap.Int("ralCode", calRalCode(resp, err)),
	)

	return httpCode, response, fields, err
}

func drainAndCloseBody(resp *http.Response, maxBytes int64) {
	if resp != nil {
		_, _ = io.CopyN(ioutil.Discard, resp.Body, maxBytes)
		_ = resp.Body.Close()
	}
}

// retry 策略
type RetryPolicy func(resp *http.Response, err error) bool

var defaultRetryPolicy RetryPolicy = func(resp *http.Response, err error) bool {
	return err != nil || resp == nil || resp.StatusCode >= 500 || resp.StatusCode == 0
}

// 重试策略
type BackOffPolicy func(attemptCount int) time.Duration

var defaultBackOffPolicy = func(attemptNum int) time.Duration { // retry immediately
	return 0
}

type ApiRequest struct {
	Req        *http.Request
	clientInit sync.Once
	HTTPClient *http.Client
	// 重试策略，可不指定，默认使用`defaultRetryPolicy`(只有在`api.yaml`中指定retry>0 时生效)
	RetryPolicy RetryPolicy
	// 重试间隔机制，可不指定，默认使用`defaultBackOffPolicy`(只有在`api.yaml`中指定retry>0 时生效)
	BackOffPolicy BackOffPolicy
}

func (entity *ApiRequest) GetRetryPolicy() RetryPolicy {
	r := defaultRetryPolicy
	if entity.RetryPolicy != nil {
		r = entity.RetryPolicy
	}
	return r
}
func (entity *ApiRequest) GetBackOffPolicy() BackOffPolicy {
	b := defaultBackOffPolicy
	if entity.BackOffPolicy != nil {
		b = entity.BackOffPolicy
	}

	return b
}

func (entity *HttpClient) Request(method string, uri string, reqBody []byte, header map[string]string, cookies map[string]string, contentType string) (data []byte, err error) {
	u := fmt.Sprintf("%s%s", entity.ApiConf.Domain, uri)

	req, err := entity.MakeRequest(method, u, strings.NewReader(string(reqBody)), header, cookies, contentType)
	if err != nil {
		entity.LogWarn("http client makeRequest error: " + err.Error())
		return nil, err
	}
	entity.LogDebugf("HttpPostJson start request: "+u, fmt.Sprintf("params", string(reqBody)))

	t := entity.beforeHttpStat(req)
	httpCode, body, fields, err := entity.httpDo(&ApiRequest{
		Req: req,
	})
	entity.afterHttpStat(req.URL.Scheme, t)

	entity.GetLog().Desugar().Debug(fmt.Sprintf("HttpPostJson end request, response code %d, body: %s", httpCode, string(body)))

	msg := "http request success"
	if err != nil {
		msg = err.Error()
	}
	entity.GetLog().Desugar().Debug(msg, fields...)

	if err != nil {
		entity.LogWarnf("PostDataJson Error [input:%v] [err:%s] [http code:%v]:\n", bytes2str(reqBody), err, httpCode)
	}
	return body, err
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func GetRequestData(requestBody interface{}) (encodeData string, err error) {
	if requestBody == nil {
		return encodeData, nil
	}

	v := url.Values{}
	if data, ok := requestBody.(map[string]string); ok {
		for key, value := range data {
			v.Add(key, value)
		}
	} else if data, ok := requestBody.(map[string]interface{}); ok {
		for key, value := range data {
			var vStr string
			switch value.(type) {
			case string:
				vStr = value.(string)
			default:
				if tmp, err := jsoniter.Marshal(value); err != nil {
					return encodeData, err
				} else {
					vStr = string(tmp)
				}
			}
			v.Add(key, vStr)
		}
	} else {
		return encodeData, errors.New("unSupport RequestBody type")
	}
	encodeData, err = v.Encode(), nil

	return encodeData, err
}

func GetUrlData(data map[string]string) (string, error) {
	v := url.Values{}
	if len(data) > 0 {
		for key, value := range data {
			v.Add(key, value)
		}
	}
	return v.Encode(), nil
}

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
