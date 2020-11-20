package daoUser

import (
    "github.com/go-opener/ctxflow/layer"

    //"time"
)

const TableDemoUser = "demoUser"

type DemoUser struct {
    Uid      uint64 `gorm:"primary_key"`
    Name     string
    Age      int32
}

func (DemoUser) TableName() string {
    return TableDemoUser
}

type DemoUserDao struct {
    layer.Dao
}

func (entity *DemoUserDao) PreUse(args ...interface{}) {
    entity.SetTable(TableDemoUser)
    entity.Dao.PreUse(args...)
}

func (entity *DemoUserDao) GetUserList() ([]DemoUser, error) {
    var result []DemoUser
    err := entity.GetDB().Find(&result).Error
    return result, err
}

func (entity *DemoUserDao) GetUserByName(name string) (DemoUser, error) {
    var result DemoUser
    err := entity.GetDB().Where("name = ?",name).First(&result).Error
    return result, err
}

