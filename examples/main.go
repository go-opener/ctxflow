package main

import (
    "examples/controller/demo"
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/go-opener/ctxflow"
    "github.com/go-opener/ctxflow/puzzle"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "os"
    "time"
)

func main() {
    // 初始化gin
    engine := gin.New()
    sLogger := getSugarLoger()
    dbClient, _ := getDBClient(sLogger)
    //设置LOG对象
    puzzle.InitConfig(puzzle.Config{
        DefaultSugaredLogger: sLogger,
        MysqlClient:dbClient,
        AppName: "demo",
        LocalIP: "127.0.0.1",
    })

    demoGroup := engine.Group("/demo")
    {
        demoGroup.POST("/testLog", ctxflow.UseController(new(demo.TestLog)))
        demoGroup.POST("/testGetUserList", ctxflow.UseController(new(demo.TestGetUserList)))
        demoGroup.POST("/testAddUser", ctxflow.UseController(new(demo.TestAddUser)))
        demoGroup.POST("/testHttpGet", ctxflow.UseController(new(demo.TestHttpGet)))
        demoGroup.POST("/panic", ctxflow.UseController(new(demo.TestPanic)))
    }
    engine.Run("0.0.0.0:8989")
}

func getDBClient(log *zap.SugaredLogger) (*gorm.DB, error) {

    client, err := gorm.Open(
        mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=True&loc=Asia%%2FShanghai",
            "root",           //user
            "",               //password
            "localhost:3306", //addr
            "demo",           //database
            "10s",            //connTimeOut
            "9500ms",         //ReadTimeOut,
            "9500ms",         //WriteTimeOut
        )),
    )

    if err != nil {
        fmt.Printf("db connect error :%+v\n", err)
        return client, err
    }

    sqlDB, err := client.DB()
    if err != nil {
        return client, err
    }

    // SetMaxIdleConns 设置空闲连接池中连接的最大数量
    sqlDB.SetMaxIdleConns(10)

    // SetMaxOpenConns 设置打开数据库连接的最大数量
    sqlDB.SetMaxOpenConns(1000)

    // SetConnMaxLifetime 设置了连接可复用的最大时间
    sqlDB.SetConnMaxLifetime(3600*time.Second)


    return client, nil
}

func getSugarLoger() *zap.SugaredLogger {
    //配置LOG文件
    fd, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
    if err != nil {
        panic("open log file error: " + err.Error())
    }
    writer := zapcore.AddSync(fd)

    var zapCore []zapcore.Core
    zapCore = append(zapCore,
        zapcore.NewCore(
            getEncoder(),
            zapcore.AddSync(writer),
            zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
                return lvl >= zapcore.DebugLevel //debug的都返回
            })))
    // core
    core := zapcore.NewTee(zapCore...)
    // 开启开发模式，堆栈跟踪
    caller := zap.AddCaller()
    // 由于之前没有DPanic，同化DPanic和Panic
    development := zap.Development()
    // 设置初始化字段
    filed := zap.Fields()
    return zap.New(core, filed, caller, development).WithOptions(zap.AddCallerSkip(1)).Sugar()
}

func getEncoder() zapcore.Encoder {
    // 公用编码器
    timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
        enc.AppendString(t.Format("2006-01-02 15:04:05"))
    }

    encoderCfg := zapcore.EncoderConfig{
        LevelKey:       "level",
        TimeKey:        "time",
        CallerKey:      "file",
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
        EncodeLevel:    zapcore.CapitalLevelEncoder,
        EncodeTime:     timeEncoder,
        EncodeDuration: zapcore.StringDurationEncoder,
    }
    return zapcore.NewJSONEncoder(encoderCfg)
}
