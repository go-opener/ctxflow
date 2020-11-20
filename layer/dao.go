package layer

import (
    "database/sql/driver"
    "github.com/jinzhu/gorm"
    "go.uber.org/zap"
    "reflect"
    "github.com/go-opener/ctxflow/puzzle"
    "fmt"
    "regexp"
    "time"
    "unicode"
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

func (entity *Dao) Print(values ...interface{}) {

    if len(values) == 0  {
        return
    }

    msg, fields := entity.gormLogKeyValueFormatter(values...)

    end := time.Now()
    fields = append(fields,zap.String("module", puzzle.GetAppName()) )
    fields = append(fields,zap.String("requestEndTime", GetFormatRequestTime(end)))
    fields = append(fields,zap.Int("ralCode", 0))
    fields = append(fields,zap.String("prot", "mysql"))

    if values[0] == "sql" && len(values) > 3 {
        startNs := end.UnixNano() - values[2].(time.Duration).Nanoseconds()
        start := time.Unix(startNs/1e9, startNs*1.0%1e9)
        fields = append(fields, zap.String("requestStartTime", GetFormatRequestTime(start)))
    }

    entity.GetLog().Desugar().Info(msg, fields...)
}

func isPrintable(s string) bool {
    for _, r := range s {
        if !unicode.IsPrint(r) {
            return false
        }
    }
    return true
}

func (entity *Dao) gormLogKeyValueFormatter(values ...interface{}) (msg string, fields []zap.Field) {
    if values[0] == "sql" {
        var sql string
        var formattedValues []string

        // sql
        for _, value := range values[4].([]interface{}) {
            indirectValue := reflect.Indirect(reflect.ValueOf(value))
            if indirectValue.IsValid() {
                value = indirectValue.Interface()
                if t, ok := value.(time.Time); ok {
                    formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
                } else if b, ok := value.([]byte); ok {
                    if str := string(b); isPrintable(str) {
                        formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
                    } else {
                        formattedValues = append(formattedValues, "'<binary>'")
                    }
                } else if r, ok := value.(driver.Valuer); ok {
                    if value, err := r.Value(); err == nil && value != nil {
                        formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
                    } else {
                        formattedValues = append(formattedValues, "NULL")
                    }
                } else {
                    switch value.(type) {
                    case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
                        formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
                    default:
                        formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
                    }
                }
            } else {
                formattedValues = append(formattedValues, "NULL")
            }
        }

        // differentiate between $n placeholders or else treat like ?
        if numericPlaceHolderRegexp.MatchString(values[3].(string)) {
            sql = values[3].(string)
            for index, value := range formattedValues {
                placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
                sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
            }
        } else {
            formattedValuesLength := len(formattedValues)
            for index, value := range sqlRegexp.Split(values[3].(string), -1) {
                sql += value
                if index < formattedValuesLength {
                    sql += formattedValues[index]
                }
            }
        }

        fields = []zap.Field{
            zap.String(TopicType, LogNameModule),
            zap.String("logId", entity.GetLogCtx().LogId),
            zap.String("requestId", entity.GetLogCtx().ReqId),
            zap.String("sql", sql),
            zap.Float64("cost", float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0),
            zap.Int64("affectedrow", values[5].(int64)),
            zap.Int("ralCode", 0),
            zap.String("prot", "mysql"),
        }

        // todo: 这里打印的日志并不代表真的成功。比如 table doesn't exist 会先打印一条日志，然后输出本sql语句
        return "mysql do success", fields
    }

    if values[0] == "log" {
        fileLineNum := values[1]
        if reflect.ValueOf(values).Kind().String() == "slice" {
            fields = []zap.Field{
                zap.String(TopicType, LogNameModule),
                zap.Reflect("file", fileLineNum),
                zap.String("logId", entity.GetLogCtx().LogId),
                zap.String("requestId", entity.GetLogCtx().ReqId),
                zap.Int("ralCode", -1),
                zap.String("prot", "mysql"),
            }

            if len(values) > 2 {
                err := values[2]
                var item reflect.Value
                if reflect.ValueOf(err).Kind() == reflect.Ptr {
                    item = reflect.ValueOf(err).Elem()
                }else{
                    item = reflect.ValueOf(err)
                }
                if !item.IsValid() {
                    return fmt.Sprintf("undecode error:%+v",values), fields
                }
                if item.MethodByName("Error").IsValid() && item.MethodByName("Error").Kind() != reflect.Invalid {
                    return err.(error).Error(), fields
                }else{
                    return fmt.Sprintf("undecode error:%+v",err), fields
                }
            }

            return fmt.Sprintf("undecode error:%+v",values), fields
        }
    }
    return
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
    return entity.db.Table(entity.GetTable()+parStr)
}

func (entity *Dao) SetDB(db *gorm.DB) {
    table := entity.GetTable()
    if table == "" {
        entity.db = db.Table(table)
    } else {
        entity.db = db
    }

    db.SetLogger(entity)
}

func (entity *Dao) SetTable(tableName string) {
    entity.tableName = tableName

}

func (entity *Dao) GetTable() string {
    return entity.tableName
}

// Update selected Fields, if attrs is an object, it will ignore default value field; if attrs is map, it will ignore unchanged field.
func (entity *Dao) Update(attrs interface{}, query interface{}, args ...interface{}) error {
    var err error
    db := entity.GetDB().Where(query, args...).Update(attrs)

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
