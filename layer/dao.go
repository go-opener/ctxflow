package layer

import (
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "reflect"
    "github.com/go-opener/ctxflow/puzzle"
    "fmt"
    "regexp"
)

// 日志重定义
var (
    sqlRegexp                = regexp.MustCompile(`\?`)
    numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

type IDao interface {
    IFlow
    GetDB() *gorm.DB
    SetDB(db *gorm.DB)
    SetTable(tableName string)
    GetTable() string
}

type Dao struct {
    Flow
    db        *gorm.DB
    tableName string
}

func (entity *Dao) Printf(msg string, args ...interface{}) {
    entity.LogInfo(fmt.Sprintf(msg,args...))
}

//默认第一个参数为db,可利用该特性批量处理事务
func (entity *Dao) PreUse(args ...interface{}) {
    if len(args) > 0 && args[0] != nil && reflect.TypeOf(args[0]).Kind() == reflect.TypeOf(&gorm.DB{}).Kind() {
        entity.SetDB(args[0].(*gorm.DB))
    } else {
        entity.SetDB(puzzle.GetDefaultGormDb())
    }
    entity.Flow.PreUse(args...)
}

func (entity *Dao) GetDB(args ...string) *gorm.DB {
    var parStr string
    if len(args) > 0{
        parStr = args[0]
    }else {
        parStr = ""
    }
    db := entity.db.WithContext(entity.GetContext()).Table(entity.GetTable()+parStr)

    entity.LogInfof("db:%+v",db.Logger)
    if !puzzle.IgnoreDefaultDBLogFormat {
        db.Logger = logger.New(entity,logger.Config{
            LogLevel: logger.Info,
        })
    }
    return db
}

func (entity *Dao) SetDB(db *gorm.DB) {
    table := entity.GetTable()
    if table == "" {
        entity.db = db.Table(table)
    } else {
        entity.db = db
    }

}

func (entity *Dao) SetTable(tableName string) {
    entity.tableName = tableName

}

func (entity *Dao) GetTable() string {
    return entity.tableName
}

// Update selected Fields, if attrs is an object, it will ignore default value field; if attrs is map, it will ignore unchanged field.
func (entity *Dao) Update(model interface{},attrs interface{}, query interface{}, args ...interface{}) error {
    var err error
    db := entity.GetDB().Model(model).Where(query, args...).Updates(attrs)

    if err = db.Error; err != nil {
        entity.LogWarnf("failed to update [tblName:%s], [query: %+v], [args: %+v], [attrs: %+v] [err:%v]", entity.GetTable(), query, args, attrs, err)
    }

    if db.RowsAffected == 0 {
        entity.LogInfof("No rows is updated.For [tblName:%s], [query: %+v], [args: %+v], [attrs: %+v] [err:%v]", entity.GetTable(), query, args, attrs, err)
    }
    return err
}

func (entity *Dao) Create(value interface{}) error {
    db := entity.GetDB()
    err := db.Create(value).Error
    if err != nil {
        entity.LogWarnf("failed to query [tableName:%s], [err: %v]", entity.GetTable(), err)
    }
    return err
}
