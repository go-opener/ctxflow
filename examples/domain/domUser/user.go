package domUser

import (
    "examples/dao/daoUser"
    "github.com/go-opener/ctxflow/layer"
)

type UserDomain struct {
    layer.Domain
}



func (entity *UserDomain) GetUserByName(name string) (daoUser.DemoUser,error) {
    entity.LogInfo("this is UserDomain log")
    userDao := entity.Use(new(daoUser.DemoUserDao)).(*daoUser.DemoUserDao)

    return userDao.GetUserByName(name)
}

