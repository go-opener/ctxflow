package layer

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "math"
    "github.com/go-opener/ctxflow/puzzle"
)

type IFlow interface {
    GetContext() *gin.Context
    SetContext(*gin.Context) *Flow
    GetArgs(idx int) interface{}
    SetArgs(idx int, arg interface{}) *Flow
    GetAllArgs() []interface{}
    Use(newFlow IFlow, args ...interface{}) interface{}
    PreUse(args ...interface{})
    GetLogCtx() *puzzle.LogCtx
    SetLogCtx(logCtx *puzzle.LogCtx) *Flow
    GetLog() *zap.SugaredLogger
    SetLog(log *zap.SugaredLogger) *Flow
}


type Flow struct {
    //zap.SugaredLogger
    ctx  *gin.Context
    log  *zap.SugaredLogger
    logS *puzzle.LogCtx
    args []interface{}
}

//参数为Use的可选参数部分
func (entity *Flow) PreUse(args ...interface{}) {
    //note 子类继承的时候使用钩子函数别忘记调用父类的方法
    //entity.XXXX.PreUse(args...)
}

//use 支持可选参数
func (entity *Flow) Use(newFlow IFlow, args ...interface{}) interface{} {
    newFlow.SetLog(entity.GetLog())
    newFlow.SetLogCtx(entity.GetLogCtx())

    newFlow.SetContext(entity.GetContext())
    for i := 0; i < int(math.Max(float64(len(args)), float64(entity.CountArgs()))); i++ {
        if len(args) > i && args[i] != nil {
            newFlow.SetArgs(i, args[i]) //优先设置参数里的args
        } else if entity.CountArgs() > i && entity.GetArgs(i) != nil {
            newFlow.SetArgs(i, entity.GetArgs(i)) //次要设置当前CtxFlow里的args
        }
    }

    newFlow.PreUse(newFlow.GetAllArgs()...)
    return newFlow
}

func (entity *Flow) GetContext() *gin.Context {
    return entity.ctx
}

func (entity *Flow) SetContext(ctx *gin.Context) *Flow {
    entity.ctx = ctx
    return entity
}

func (entity *Flow) GetLogCtx() *puzzle.LogCtx {
    return entity.logS
}

func (entity *Flow) SetLogCtx(logCtx *puzzle.LogCtx) *Flow {
    entity.logS = logCtx
    return entity
}

func (entity *Flow) GetLog() *zap.SugaredLogger {
    return entity.log
}

func (entity *Flow) SetLog(log *zap.SugaredLogger) *Flow {
    entity.log = log
    return entity
}

func (entity *Flow) genArgs(idx int) bool {
    if idx >= 10 { //最多存储10条
        return false
    }
    for len(entity.args)-1 < idx {
        entity.args = append(entity.args, nil)
    }
    return true
}

func (entity *Flow) GetArgs(idx int) interface{} {
    if entity.CountArgs() > idx {
        return entity.args[idx]
    } else {
        return nil
    }
}

func (entity *Flow) CountArgs() int {
    return len(entity.args)
}

func (entity *Flow) SetArgs(idx int, arg interface{}) *Flow {
    if entity.genArgs(idx) {
        entity.args[idx] = arg
    }

    return entity
}

func (entity *Flow) GetAllArgs() []interface{} {
    return entity.args
}

// 提供给业务使用的server log 日志打印方法
func (entity *Flow) LogDebug(args ...interface{}) {
    entity.log.Debug(args...)
}

func (entity *Flow) LogDebugf(format string, args ...interface{}) {
    entity.log.Debugf(format, args...)
}

func (entity *Flow) LogInfo(args ...interface{}) {
    entity.log.Info(args...)
}

func (entity *Flow) LogInfof(format string, args ...interface{}) {
    entity.log.Infof(format, args...)
}

func (entity *Flow) LogWarn(args ...interface{}) {
    entity.log.Warn(args...)
}

func (entity *Flow) LogWarnf(format string, args ...interface{}) {
    entity.log.Warnf(format, args...)
}

func (entity *Flow) LogError(args ...interface{}) {
    entity.log.Error(args...)
}

func (entity *Flow) LogErrorf(format string, args ...interface{}) {
    entity.log.Errorf(format, args...)
}
