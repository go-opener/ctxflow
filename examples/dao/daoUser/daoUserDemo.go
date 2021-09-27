package daoUser

import (
    "github.com/go-opener/ctxflow/layer"

    //"time"
)

const TableDemoUser = "demoUser"

type DemoUserDao struct {
    layer.Dao
    Uid  uint64 `gorm:"primary_key"`
    Name string
    Age  int32
}

func (DemoUserDao) TableName() string {
    return TableDemoUser
}

func (entity *DemoUserDao) PreUse(args ...interface{}) {
    entity.SetModel(entity)
    entity.Dao.PreUse(args...)
}

func (entity *DemoUserDao) GetUserList() ([]DemoUserDao, error) {
    var result []DemoUserDao
    err := entity.GetDB().Find(&result).Error
    return result, err
}

func (entity *DemoUserDao) GetUserByName(name string) (DemoUserDao, error) {
    var result DemoUserDao
    err := entity.GetDB().Where("name = ?", name).First(&result).Error
    return result, err
}
