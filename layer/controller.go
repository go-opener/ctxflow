package layer

import (
    "errors"
    "fmt"
    "github.com/gin-gonic/gin"
    jsoniter "github.com/json-iterator/go"
    "github.com/mitchellh/mapstructure"
    "gopkg.in/go-playground/validator.v9"
    "net/http"
    "reflect"
    "unsafe"
    "github.com/go-opener/ctxflow/v2/puzzle"
)



var (
	//StackLogger func(ctx *gin.Context, err error)
    ErrMsgMap                  map[int]string
    NmqResponseStatusCodeError = 500
    NmqRetryError              = 34002
)


type IController interface {
    IFlow
    BindParam(req interface{}) bool
    Action()
    RenderJsonFail(err error)
    RenderJsonSucc(data interface{})
    RenderJsonAbort(err error)
    RenderHttpError(errNo int, errDetail ...interface{})
}

type Controller struct {
    Flow
}

func (entity *Controller) Action() {

}
func (entity *Controller) BindParam(req interface{}) bool {
    err := entity.BindParamError(req)
    if err != nil {
        return false
    }
    return true
}

func (entity *Controller) BindParamError(req interface{}) error {

    var defaultMap map[string]interface{}
    var outMap map[string]interface{}

    mapstructure.Decode(req,&defaultMap)

    ctx := entity.GetContext()
    if errParams := ctx.ShouldBindJSON(req); errParams != nil {
        entity.LogWarnf("param error: %+v", errParams)
        return errParams
    }

    mapstructure.Decode(req,&outMap)

    validate := validator.New()
    validErr := validate.Struct(req)
    if validErr != nil {
        entity.LogWarnf("param error: %+v", validErr)
        return validErr
    }

    for key,val := range defaultMap{

        if reflect.TypeOf(outMap[key]).Kind() != reflect.Ptr {
            continue
        }

        if reflect.ValueOf(outMap[key]).IsNil()  && !reflect.ValueOf(val).IsNil(){
           entity.LogWarnf("param error,invalid input: %+v", key)
           return errors.New(fmt.Sprintf("param error,invalid input:%+v",key))
        }
    }

    return nil
}

type BaseError interface {

}

func ErrorToRanderJson(err error) puzzle.DefaultRender{
    var renderJson puzzle.DefaultRender

    if err == nil {
        renderJson.ErrNo = 0
        renderJson.ErrMsg = ""
    }else if reflect.TypeOf(err).Kind() == reflect.Ptr {
        n := reflect.ValueOf(err).Elem().FieldByName("ErrNo")
        if n.Kind() != reflect.Invalid {
            renderJson.ErrNo = int(n.Int())
        }else{
            renderJson.ErrNo = 9999
        }

        m := reflect.ValueOf(err).Elem().FieldByName("ErrMsg")
        s := reflect.ValueOf(err).Elem().FieldByName("ErrStr")
        if m.Kind() != reflect.Invalid {
            renderJson.ErrMsg = m.String()
        }else if s.Kind() != reflect.Invalid {
            renderJson.ErrMsg = s.String()
        }else {
            renderJson.ErrMsg = err.Error()
        }
    }else if reflect.TypeOf(err).Kind() == reflect.Struct {
        n := reflect.ValueOf(err).FieldByName("ErrNo")
        if n.Kind() != reflect.Invalid {
            renderJson.ErrNo = int(n.Int())
        }else{
            renderJson.ErrNo = 9999
        }

        m := reflect.ValueOf(err).FieldByName("ErrMsg")
        s := reflect.ValueOf(err).FieldByName("ErrStr")
        if m.Kind() != reflect.Invalid {
            renderJson.ErrMsg = m.String()
        }else if s.Kind() != reflect.Invalid {
            renderJson.ErrMsg = s.String()
        }else {
            renderJson.ErrMsg = err.Error()
        }
    }
    return renderJson
}

func (entity *Controller) RenderJsonFail(err error) {
    // 打印错误栈
    //if StackLogger != nil {
    //    StackLogger(entity.GetContext(), err)
    //}

    entity.Info(err)
    renderJson := ErrorToRanderJson(err)
    renderJson.Data = gin.H{}
    entity.GetContext().JSON(http.StatusOK, renderJson)
}

func (entity *Controller) RenderJsonSucc(data interface{}) {
    ctx := entity.GetContext()
    renderJson := puzzle.DefaultRender{0, "succ", data}
    ctx.JSON(http.StatusOK, renderJson)
}

func (entity *Controller) RenderJsonAbort(err error) {
    entity.Info(err)
    renderJson := ErrorToRanderJson(err)
    entity.GetContext().AbortWithStatusJSON(http.StatusOK, renderJson)
}

func (entity *Controller) RenderHttpError(errNo int, errDetail ...interface{}) {
    var errStr string = ""
    if ErrMsgMap !=nil{
        if s, ok := ErrMsgMap[errNo]; ok {
            errStr = s
        }
    }

    entity.Info(errDetail)
    body := gin.H{"errNo": errNo, "errStr": errStr, "data": map[string]interface{}{}, "errDetail": errDetail}
    data, _ := jsoniter.Marshal(body)
    entity.GetContext().String(NmqResponseStatusCodeError, toString(data))
}

func toString(s []byte) string {
    return *(*string)(unsafe.Pointer(&s))
}