package puzzle

import "gorm.io/gorm"
var MysqlClient *gorm.DB
//deprecated
func SetDefaultGormDb(db *gorm.DB){
    MysqlClient = db
}

func GetDefaultGormDb()*gorm.DB{
    return MysqlClient
}