package puzzle

import "go.uber.org/zap"
import "gorm.io/gorm"

var IgnoreDefaultDBLogFormat bool = false

type Config struct {
    DefaultSugaredLogger *zap.SugaredLogger
    MysqlClient *gorm.DB
    KafkaClient IKafka
    RedisClient IRedis
    IgnoreDefaultDBLogFormat bool  //是否使用默认的数据库log输出格式，如果有些框架会自定义输出格式，此项选false
    AppName string
    LocalIP string
}

func InitConfig(config Config){
    if config.DefaultSugaredLogger != nil {
        DefaultSugaredLogger = config.DefaultSugaredLogger
    }
    if config.KafkaClient != nil {
        KafkaClient = config.KafkaClient
    }
    if config.MysqlClient != nil {
        MysqlClient = config.MysqlClient
    }
    if config.RedisClient != nil {
        RedisClient = config.RedisClient
    }
    if config.IgnoreDefaultDBLogFormat == true {
        IgnoreDefaultDBLogFormat = config.IgnoreDefaultDBLogFormat
    }
    AppName = config.AppName
    LocalIp = config.LocalIP

}
