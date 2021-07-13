package adapter

import (
    "encoding/json"
    "errors"
    "examples/adapter/httpClient"
    "fmt"
    "github.com/go-opener/ctxflow/layer"
    "github.com/go-opener/ctxflow/puzzle"
    jsoniter "github.com/json-iterator/go"
    "net/url"
)

type DefaultAdapter struct {
    layer.Flow
    *httpClient.HttpClient
}

func (entity *DefaultAdapter) PreUse(args ...interface{}) {
    entity.HttpClient = entity.Use(new(httpClient.HttpClient)).(*httpClient.HttpClient)
    entity.HttpClient.ApiConf = args[0].(httpClient.ApiConf)
    entity.Flow.PreUse(args...)
}

func (entity *DefaultAdapter) Request(method string, uri string, reqBody []byte, header map[string]string, cookies map[string]string, contentType string) (data []byte, err error) {
    return entity.HttpClient.Request(method , uri, reqBody, header, cookies, contentType)
}


func (entity *DefaultAdapter) PostDataJson(uri string, params map[string]interface{}, header map[string]string) (data []byte, err error) {
   reqBody, err := jsoniter.Marshal(params)
   if err != nil {
       entity.LogWarn("http client make data error: " + err.Error())
       return nil, err
   }
   return entity.Request("POST", uri, reqBody, header, nil, "application/json")
}

func (entity *DefaultAdapter) PostData(uri string, params map[string]interface{}) (data []byte, err error) {

   reqBody, err := httpClient.GetRequestData(params)

   if err != nil {
       entity.LogWarn("http client make data error: " + err.Error())
       return nil, err
   }

   return entity.Request("POST", uri, httpClient.Str2bytes(reqBody), nil, nil, "application/x-www-form-urlencoded")
}

func (entity *DefaultAdapter) HttpPostJSON(uri string, params map[string]interface{}, header map[string]string) (data []byte, err error) {
   reqBody, err := json.Marshal(params)
   if err != nil {
       entity.LogWarn("http client make data error: " + err.Error())
       return nil, err
   }
   return entity.Request("POST", uri, reqBody, header, nil, "application/json")
}

func (entity *DefaultAdapter) HttpPost(uri string, params map[string]string, header map[string]string) (data []byte, err error) {
   reqBody, err := httpClient.GetUrlData(params)
   if err != nil {
       entity.LogWarn("http client make data error: " + err.Error())
       return nil, err
   }

   return entity.Request("POST", uri, httpClient.Str2bytes(reqBody), header, nil, "application/x-www-form-urlencoded")
}

func (entity *DefaultAdapter) HttpGet(uri string, params map[string]string, header map[string]string) (data []byte, err error) {
   reqBody, err := httpClient.GetUrlData(params)
   if err != nil {
       entity.LogWarn("http client make data error: " + err.Error())
       return nil, err
   }

   return entity.Request("GET", uri, httpClient.Str2bytes(reqBody), header, nil, "application/x-www-form-urlencoded")
}

func (entity *DefaultAdapter) RalPost(uri string, params map[string]interface{}, header map[string]string) (*puzzle.DefaultRender, error) {

   reqBody, err := json.Marshal(params)
   if err != nil {
       entity.LogWarn("http client make data error: " + err.Error())
       return nil, err
   }
   rst, err := entity.Request("POST", uri, reqBody, header, nil, "application/json")
   if err != nil {
       entity.LogWarnf("Http Post Error [input:%v] [err:%s] [http code:%v]:\n", params, err, rst)

       return nil, err
   }

   var resMap puzzle.DefaultRender

   err = jsoniter.Unmarshal(rst, &resMap)
   if err != nil {
       entity.LogWarnf(" res unmarshal error [res:%s] [err:%s]", rst, err.Error())
       return nil, err
   }

   if resMap.ErrNo != 0 {
       entity.LogWarnf("request data error [res:%s]", resMap)
       return nil, errors.New("RalPost 获取资源数据错误")
   }

   return &resMap, nil
}

func (entity *DefaultAdapter) PostForm(uri string, params map[string]interface{}, header *map[string]string) ([]byte, error) {
   ret := make(map[string]string, len(params))
   for k, v := range params {
       ret[k] = fmt.Sprint(v)
   }

   v := url.Values{}

   for key, value := range ret {
       v.Add(key, value)
   }

   return entity.Request("POST", uri, httpClient.Str2bytes(v.Encode()), nil, nil, "application/x-www-form-urlencoded")
}
