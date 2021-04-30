package demo

import (
    "examples/dto/dtoUser"
    "examples/service/svUser"
    "github.com/apache/thrift/lib/go/thrift"
    "github.com/go-opener/ctxflow/v2/layer"
)

type TestAddUser struct {
    layer.Controller
}

func (entity *TestAddUser) Action() {
    entity.LogInfo("test getUser start")

    req := dtoUser.AddUserReq{
        Age: thrift.Int32Ptr(10),//给age赋予默认值
    }

    if err :=entity.BindParamError(&req);err != nil{
        entity.RenderJsonFail(err)
        return
    }

    userService := entity.Use(new(svUser.UserService)).(*svUser.UserService)
    err := userService.AddUser(&req)
    if err != nil {
        entity.RenderJsonFail(err)
        return
    }
    entity.RenderJsonSucc("success")
}
