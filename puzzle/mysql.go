package puzzle

import "github.com/jinzhu/gorm"
var MysqlClient *gorm.DB
func SetDefaultGormDb(db *gorm.DB){
    MysqlClient = db
}

func GetDefaultGormDb()*gorm.DB{
    return MysqlClient
}