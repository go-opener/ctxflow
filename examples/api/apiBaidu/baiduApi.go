package apiBaidu

import (
    "github.com/go-opener/ctxflow/layer"
    jsoniter "github.com/json-iterator/go"
)

type Baidu struct {
    layer.Api
}

func (entity *Baidu)PreUse(args ...interface{}){
    entity.ApiConf = layer.ApiConf{
        Service: "baidu",
        Domain: "http://tieba.baidu.com",
    }
    entity.Api.PreUse(args...)
}

// 获取demo内容
func (entity *Baidu)GetLiveStat() (map[string]interface{}, error) {
    params := map[string]string{
        "fids":"",
        "_t":"1605271642810",
    }

    bytes, err := entity.HttpGet("/show/getlivestat", params, nil)

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
