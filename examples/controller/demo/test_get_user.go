package demo

import (
    "examples/dao/daoUser"
    //"fmt"
    "github.com/go-opener/ctxflow/layer"
)

type TestGetUserList struct {
    layer.Controller
}

func (entity *TestGetUserList) Action() {
    entity.LogInfo("test getUser start")
    user := entity.Use(new(daoUser.DemoUserDao)).(*daoUser.DemoUserDao)
    list,_ :=user.GetUserList()
    entity.RenderJsonSucc(list)
}
