package ctxflow

import (
    "bytes"
    "encoding/json"
    "errors"
    "runtime"

    "github.com/gin-gonic/gin"

    "github.com/spf13/cobra"
    "go.uber.org/zap"
    "io/ioutil"
    "reflect"

    "github.com/go-opener/ctxflow/layer"
    "github.com/go-opener/ctxflow/lib/mcpack"
    "github.com/go-opener/ctxflow/puzzle"
)

const (
    NMQ_ERR_STOP     = 5001 //参数错误，ticker停止执行
    NMQ_ERR_CONTINUE = 500  //处理失败，ticker继续执行
    NMQ_SUC_CONTINUE = 200  //处理成功，ticker继续执行
)

func Slave(src interface{}) interface{} {
    typ := reflect.TypeOf(src)
    if typ.Kind() == reflect.Ptr { //如果是指针类型
        typ = typ.Elem()                          //获取源实际类型(否则为指针类型)
        dst := reflect.New(typ).Elem()            //创建对象
        b, _ := json.Marshal(src)                 //导出json
        json.Unmarshal(b, dst.Addr().Interface()) //json序列化
        return dst.Addr().Interface()             //返回指针
    } else {
        dst := reflect.New(typ).Elem()            //创建对象
        b, _ := json.Marshal(src)                 //导出json
        json.Unmarshal(b, dst.Addr().Interface()) //json序列化
        return dst.Interface()                    //返回值
    }
}

func UseController(controller layer.IController) func(ctx *gin.Context) {
    return func(ctx *gin.Context) {
        ctl := Slave(controller).(layer.IController)

        InitFlow(ctx,ctl)

        defer NoPanicContorller(ctl)

        ctl.PreUse()
        ctl.Action()
    }
}

func NoPanicContorller(ctl layer.IController){
    if err:=recover();err!=nil{
        stack := PanicTrace(4) //4KB
        ctl.GetLog().Errorf("[controller panic]:%+v,stack:%s",err,string(stack))
        ctl.RenderJsonFail(errors.New("异常错误！"))
    }
}

func NoPanic(flow layer.IFlow){
    if err:=recover();err!=nil{
        stack := PanicTrace(4) //4KB
        flow.GetLog().Errorf("[controller panic]:%+v,stack:%s",err,string(stack))
    }
}

func StackTrace() []byte {
    e := []byte("\ngoroutine ")
    line := []byte("\n")
    stack := make([]byte, 4<<10) //4KB
    length := runtime.Stack(stack, true)
    stack = stack[0:length]
    start := bytes.Index(stack, line) + 1
    stack = stack[start:]
    end := bytes.LastIndex(stack, line)
    if end != -1 {
        stack = stack[:end]
    }
    end = bytes.Index(stack, e)
    if end != -1 {
        stack = stack[:end]
    }
    stack = bytes.TrimRight(stack, "\n")

    return stack
}

func PanicTrace(kb int) []byte {
    s := []byte("/src/runtime/panic.go")
    e := []byte("\ngoroutine ")
    line := []byte("\n")
    stack := make([]byte, kb<<10) //4KB
    length := runtime.Stack(stack, true)
    start := bytes.Index(stack, s)
    stack = stack[start:length]
    start = bytes.Index(stack, line) + 1
    stack = stack[start:]
    end := bytes.LastIndex(stack, line)
    if end != -1 {
        stack = stack[:end]
    }
    end = bytes.Index(stack, e)
    if end != -1 {
        stack = stack[:end]
    }
    stack = bytes.TrimRight(stack, "\n")
    return stack
}

func UseTask(task layer.ITask) func(cmd *cobra.Command, args []string) {
    return func(cmd *cobra.Command, args []string) {
        ctx := &gin.Context{}
        task2 := Slave(task).(layer.ITask)

        InitFlow(ctx,task2)

        task2.PreUse()
        task2.Run(args)
    }
}

func MakeFlow(ctx *gin.Context) *layer.Flow{
    if ctx == nil {
        ctx = &gin.Context{}
    }
    flow := new(layer.Flow)
    InitFlow(ctx,flow)
    return flow
}

func InitFlow(ctx *gin.Context,flow layer.IFlow){
    flow = flow.SetContext(ctx)
    logCtx := puzzle.LogCtx{
        LogId:   puzzle.GetLogID(ctx),
        ReqId:   puzzle.GetRequestID(ctx),
        AppName: puzzle.GetAppName(),
        LocalIp: puzzle.GetLocalIp(),
    }
    flow.SetLogCtx(&logCtx)
    flow.SetLog(puzzle.GetDefaultSugaredLogger().With(
        zap.String("logId", logCtx.LogId),
        zap.String("requestId", logCtx.ReqId),
        zap.String("module", logCtx.AppName),
        zap.String("localIp", logCtx.LocalIp),
    ))
}

func UseKFKConsumer(consumer layer.IConsumer) func(cmd *cobra.Command, args []string) {
    return func(cmd *cobra.Command, args []string) {
        ctx := &gin.Context{}
        consumer2 := Slave(consumer).(layer.IConsumer)
        InitFlow(ctx,consumer2)
        consumer2.PreUse()
        consumer2.Run(args)
    }
}

func UseNMQ(nmqMap map[string]reflect.Type) func(ctx *gin.Context) {
    return func(ctx *gin.Context) {
        controller := new(layer.Flow).SetContext(ctx).Use(new(layer.Controller)).(*layer.Controller)

        InitFlow(ctx,controller)

        cmdNo := ctx.Query("cmdno")
        if cmdNo == "" {
            controller.LogWarnf("[nmqservice] [commit] [param error] [cmdNo is empty]")
            controller.RenderJsonSucc(map[string]int{"status": NMQ_SUC_CONTINUE})
            return
        }

        nmqType, ok := nmqMap[cmdNo]
        if !ok {
            controller.LogErrorf("[nmqservice] [commit] [callback func is not exist] [CmdNo:%s] [Nmqs:%v]", cmdNo, nmqMap)
            controller.RenderJsonSucc(map[string]int{"status": NMQ_SUC_CONTINUE})
            return
        }

        var consumer = reflect.New(nmqType).Elem().Addr().Interface().(layer.INMQConsumer)

        rawData, _ := ioutil.ReadAll(ctx.Request.Body)
        if err := mcpack.Unmarshal(rawData, consumer); err != nil {
            controller.LogErrorf("[nmqservice] [commit] [mcpack param unmashall error] [CmdNo:%v] [Err:%v] [data:%s]", cmdNo, err, rawData)
            controller.RenderJsonSucc(map[string]int{"status": NMQ_ERR_CONTINUE})
            return
        }
        param, err := json.Marshal(consumer)
        controller.LogInfof("[nmqservice] [commit] [Cmdno:%s] [Transid:%s] [RequestParam:%s] [Err:%v]", cmdNo, ctx.Query("transid"), string(param), err)

        InitFlow(ctx,consumer)

        consumer.PreUse()
        //do
        d1, proErr := consumer.Process()

        if proErr == nil {
            controller.RenderJsonSucc(d1)
            return
        } else {
            renderJson := layer.ErrorToRanderJson(proErr)
            if renderJson.ErrNo == layer.NmqRetryError {
                controller.RenderHttpError(renderJson.ErrNo, proErr)
            } else {
                controller.RenderJsonFail(proErr)
            }
        }

    }
}
