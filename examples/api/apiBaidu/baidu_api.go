package apiBaidu

import (
    "examples/adapter"
    "examples/adapter/httpClient"
    "github.com/go-opener/ctxflow/layer"
    jsoniter "github.com/json-iterator/go"
)

type Baidu struct {
    layer.Api
    defaultClient *adapter.DefaultAdapter
}

func (entity *Baidu)PreUse(args ...interface{}){

    entity.defaultClient = entity.Use(new(adapter.DefaultAdapter),httpClient.ApiConf{
        Service: "baidu",
        Domain: "http://tieba.baidu.com",
    }).(*adapter.DefaultAdapter)

    entity.Api.PreUse(args...)
}

// 获取demo内容
func (entity *Baidu)GetLiveStat() (map[string]interface{}, error) {
    params := map[string]string{
        "fids":"",
        "_t":"1605271642810",
    }


    reqBody, err := httpClient.GetRequestData(params)

    if err != nil {
        entity.LogWarn("http client make data error: " + err.Error())
        return nil, err
    }

    bytes,err := entity.defaultClient.Request("POST", "/show/getlivestat", httpClient.Str2bytes(reqBody), nil, nil, "application/x-www-form-urlencoded")


    if err != nil {
        entity.LogWarnf("request getexaminfo error: %v", err)
        return nil, err
    }
    result := make(map[string]interface{})
    err = jsoniter.Unmarshal(bytes, &result)
    if err != nil {
        entity.LogWarnf("getexaminfo json unmarshal error: %v", err)
        return nil, err
    }

    return result, nil
}

// 获取demo内容
func (entity *Baidu)TestHttpGet() (map[string]interface{}, error) {
    params := map[string]string{
        "fids":"",
        "_t":"1605271642810",
    }


    bytes,err := entity.defaultClient.HttpGet("/show/getlivestat", params, nil)


    if err != nil {
        entity.LogWarnf("request getexaminfo error: %v", err)
        return nil, err
    }
    result := make(map[string]interface{})
    err = jsoniter.Unmarshal(bytes, &result)
    if err != nil {
        entity.LogWarnf("getexaminfo json unmarshal error: %v", err)
        return nil, err
    }

    return result, nil
}