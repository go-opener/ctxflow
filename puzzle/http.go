package puzzle

import (
    "github.com/gin-gonic/gin"
)

type DefaultRender struct {
    ErrNo  int         `json:"errNo"`
    ErrMsg string      `json:"errStr"`
    Data   interface{} `json:"data"`
}

type IHTTPClient interface {
    PostDataJson(ctx *gin.Context,uri string, params map[string]interface{}, header map[string]string) (data []byte, err error)
    PostData(ctx *gin.Context,uri string, params map[string]interface{}) (data []byte, err error)
    HttpPostJSON(ctx *gin.Context,uri string, params map[string]interface{}, header map[string]string) (data []byte, err error)
    HttpPost(ctx *gin.Context,uri string, params map[string]string, header map[string]string) (data []byte, err error)
    HttpGet(ctx *gin.Context,uri string, params map[string]string, header map[string]string) (data []byte, err error)
    RalPost(ctx *gin.Context,uri string, params map[string]interface{}, header map[string]string) (*DefaultRender, error)
}

var HTTPClient IHTTPClient

func SetHTTPClient(client IHTTPClient) {
    HTTPClient = client
}

func GetHTTPClient() IHTTPClient {
    return HTTPClient
}
