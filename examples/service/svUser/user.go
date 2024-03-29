package svUser

import (
    "examples/dao/daoUser"
    "examples/data/dsUser"
    "examples/dto/dtoUser"
    "github.com/go-opener/ctxflow/layer"
    "github.com/go-opener/ctxflow/puzzle"
    "gorm.io/gorm"
)

type UserService struct {
    layer.Service
}

func (entity *UserService) AddUser(req *dtoUser.AddUserReq) error {
    entity.LogInfof("this is a Service log,service name:%+v","UserService")

    //db关联其他模块，统一提交事务或者回滚
    db := puzzle.GetDefaultGormDb().Begin()
    //use方法的第二个参数可选，可以是db也可以是其他。如果设置为某个DB，则被这个DB关联了事务
    userRepo := entity.Use(new(dsUser.UserRepository),db).(*dsUser.UserRepository)
    usr,err:=userRepo.GetUserByName(req.Name)

    if err != gorm.ErrRecordNotFound {
        entity.LogWarn("用户已存在:%+v",usr)
        db.Rollback()
        return err
    }

    //use方法的第二个参数可选，可以是db也可以是其他。如果设置为某个DB，则被这个DB关联了事务
    userDao := entity.Use(new(daoUser.DemoUserDao),db).(*daoUser.DemoUserDao)
    err = userDao.Create(&daoUser.DemoUserDao{
        Name: req.Name,
        Age:*req.Age,
    })

    if err != nil {
        entity.LogWarn("创建失败:%+v",usr)
        db.Rollback()
        return err
    }
    db.Commit()

    return nil
}

