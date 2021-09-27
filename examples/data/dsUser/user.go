package dsUser

import (
    "examples/dao/daoUser"
    "github.com/go-opener/ctxflow/layer"
)

type UserRepository struct {
    layer.DataSet
}



func (entity *UserRepository) GetUserByName(name string) (daoUser.DemoUserDao,error) {
    entity.LogInfo("this is UserDomain log")
    userDao := entity.Use(new(daoUser.DemoUserDao)).(*daoUser.DemoUserDao)

    return userDao.GetUserByName(name)
}

