package puzzle

import "github.com/gin-gonic/gin"

type IRedis interface {
    Do(ctx *gin.Context, commandName string, args ...interface{}) (reply interface{}, err error)
}

var RedisClient IRedis

//deprecated
func SetRedisClient(client IRedis) {
    RedisClient = client
}

func GetRedisClient() IRedis {
    return RedisClient
}
