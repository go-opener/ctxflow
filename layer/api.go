package layer

import (
    "encoding/json"
    "errors"
    "fmt"
    jsoniter "github.com/json-iterator/go"

    "net/url"

    "unsafe"
    "github.com/go-opener/ctxflow/v2/puzzle"
)



type IApi interface {
    IFlow
    PostDataJson(uri string, params map[string]interface{}, header map[string]string) (data []byte, err error)
    PostData(uri string, params map[string]interface{}) (data []byte, err error)
    HttpPostJSON(uri string, params map[string]interface{}, header map[string]string) (data []byte, err error)
    HttpPost(uri string, params map[string]string, header map[string]string) (data []byte, err error)
    HttpGet(uri string, params map[string]string, header map[string]string) (data []byte, err error)
    RalPost(uri string, params map[string]interface{}, header map[string]string) (*puzzle.DefaultRender, error)
}

type Api struct {
    Flow
    http puzzle.IHTTPClient
}

func (entity *Api) PreUse(args ...interface{}) {
    entity.Flow.PreUse(args...)
}

func (entity *Api) SetHTTP(http puzzle.IHTTPClient) {
    entity.http = http
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

func str2bytes(s string) []byte {
    x := (*[2]uintptr)(unsafe.Pointer(&s))
    h := [3]uintptr{x[0], x[1], x[1]}
    return *(*[]byte)(unsafe.Pointer(&h))
}

func (entity *Api) PostDataJson(uri string, params map[string]interface{}, header map[string]string) (data []byte, err error) {
    reqBody, err := json.Marshal(params)
    if err != nil {
        entity.LogWarn("http client make data error: " + err.Error())
        return nil, err
    }
    return entity.http.Request("POST", uri, reqBody, header, nil, "application/json")
}

func (entity *Api) PostData(uri string, params map[string]interface{}) (data []byte, err error) {

    reqBody, err := GetRequestData(params)

    if err != nil {
        entity.LogWarn("http client make data error: " + err.Error())
        return nil, err
    }

    return entity.http.Request("POST", uri, str2bytes(reqBody), nil, nil, "application/x-www-form-urlencoded")
}

func (entity *Api) HttpPostJSON(uri string, params map[string]interface{}, header map[string]string) (data []byte, err error) {
    reqBody, err := json.Marshal(params)
    if err != nil {
        entity.LogWarn("http client make data error: " + err.Error())
        return nil, err
    }
    return entity.http.Request("POST", uri, reqBody, header, nil, "application/json")
}

func (entity *Api) HttpPost(uri string, params map[string]string, header map[string]string) (data []byte, err error) {
    reqBody, err := GetUrlData(params)
    if err != nil {
        entity.LogWarn("http client make data error: " + err.Error())
        return nil, err
    }

    return entity.http.Request("POST", uri, str2bytes(reqBody), header, nil, "application/x-www-form-urlencoded")
}

func (entity *Api) HttpGet(uri string, params map[string]string, header map[string]string) (data []byte, err error) {
    reqBody, err := GetUrlData(params)
    if err != nil {
        entity.LogWarn("http client make data error: " + err.Error())
        return nil, err
    }

    return entity.http.Request("GET", uri, str2bytes(reqBody), header, nil, "application/x-www-form-urlencoded")
}

func (entity *Api) RalPost(uri string, params map[string]interface{}, header map[string]string) (*puzzle.DefaultRender, error) {

    reqBody, err := json.Marshal(params)
    if err != nil {
        entity.LogWarn("http client make data error: " + err.Error())
        return nil, err
    }
    rst, err := entity.http.Request("POST", uri, reqBody, header, nil, "application/json")
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

func (entity *Api) PostForm(uri string, params map[string]interface{}, header *map[string]string) ([]byte, error) {
    ret := make(map[string]string, len(params))
    for k, v := range params {
        ret[k] = fmt.Sprint(v)
    }

    v := url.Values{}

    for key, value := range ret {
        v.Add(key, value)
    }

    return entity.http.Request("POST", uri, str2bytes(v.Encode()), nil, nil, "application/x-www-form-urlencoded")
}



