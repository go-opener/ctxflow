package pack

import (
    "fmt"
    libRedis "github.com/gomodule/redigo/redis"
    jsoniter "github.com/json-iterator/go"
    "github.com/pkg/errors"
    "math"
    "strconv"
    "strings"
    "github.com/go-opener/ctxflow/v2/layer"
    "github.com/go-opener/ctxflow/v2/puzzle"
)

const (
    _CHUNK_SIZE = 32
)

type IRedis interface {
    layer.IFlow
}

type Redis struct {
    layer.Flow
    RedisClient puzzle.IRedis
}

//默认第一个参数为db,可利用该特性批量处理事务
func (entity *Redis) PreUse(args ...interface{}) {
    entity.RedisClient = puzzle.GetRedisClient()
    entity.Flow.PreUse(args...)
}

//将interface转为string
func ToStr(value interface{}) (s string) {
    switch v := value.(type) {
    case bool:
        s = strconv.FormatBool(v)
    case float32:
        s = strconv.FormatFloat(float64(v), 'f', -1, 32)
    case float64:
        s = strconv.FormatFloat(v, 'f', -1, 64)
    case int:
        s = strconv.FormatInt(int64(v), 10)
    case int8:
        s = strconv.FormatInt(int64(v), 10)
    case int16:
        s = strconv.FormatInt(int64(v), 10)
    case int32:
        s = strconv.FormatInt(int64(v), 10)
    case int64:
        s = strconv.FormatInt(v, 10)
    case uint:
        s = strconv.FormatUint(uint64(v), 10)
    case uint8:
        s = strconv.FormatUint(uint64(v), 10)
    case uint16:
        s = strconv.FormatUint(uint64(v), 10)
    case uint32:
        s = strconv.FormatUint(uint64(v), 10)
    case uint64:
        s = strconv.FormatUint(v, 10)
    case string:
        s = v
    case []byte:
        s = string(v)
    default:
        s = fmt.Sprintf("%v", v)
    }
    return s
}

func (entity *Redis) SetDebug(commandName string, args ...interface{}){
    sv := entity.GetContext().Query("_debug")
    if sv == "yayazhongtai"{
        debug := commandName
        for _,arg := range args{
            debug += " " + ToStr(arg)
        }
        if v,ok :=entity.GetContext().Get("debug");ok{
            entity.GetContext().Set("debug",append(v.([]string),debug))
        }else{
            entity.GetContext().Set("debug",[]string{debug})
        }
    }
}

func (entity *Redis) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
    entity.SetDebug(commandName,args...)
    return entity.RedisClient.Do(entity.GetContext(), commandName, args...)
}

func (entity *Redis) Expire(key string, time int64, opt ...int) (bool, error) {
    if len(opt) > 0 {
        noCache := opt[0]
        if noCache == 1 {
            return false, nil
        }
    }
    return libRedis.Bool(entity.Do("EXPIRE", key, time))
}

func (entity *Redis) Exists(key string) (bool, error) {
    return libRedis.Bool(entity.Do("EXISTS", key))
}

func (entity *Redis) Del(keys ...interface{}) (int64, error) {
    return libRedis.Int64(entity.Do("DEL", keys...))
}

func (entity *Redis) Ttl(key string) (int64, error) {
    return libRedis.Int64(entity.Do("TTL", key))
}

func (entity *Redis) Pttl(key string) (int64, error) {
    return libRedis.Int64(entity.Do("PTTL", key))
}

func (entity *Redis) HSet(key, field string, val interface{}, opt ...int) (int, error) {
    if len(opt) > 0 {
        noCache := opt[0]
        if noCache == 1 {
            return 0, nil
        }
    }
    valStr := parseToString(val)
    return libRedis.Int(entity.Do("HSET", key, field, valStr))
}

func parseToString(value interface{}) string {
    switch value.(type) {
    case string:
        return value.(string)
    default:
        b, e := jsoniter.Marshal(value)
        if e != nil {
            return ""
        }
        return string(b)
    }
}

func (entity *Redis) HGet(key, field string, opt ...int) ([]byte, error) {
    if len(opt) > 0 {
        noCache := opt[0]
        if noCache == 1 {
            return nil, nil
        }
    }
    if res, err := libRedis.Bytes(entity.Do("HGET", key, field)); err == libRedis.ErrNil {
        return nil, nil
    } else {
        return res, err
    }
}

func (entity *Redis) HMGet(key string, fields ...string) ([][]byte, error) {
    //1.初始化返回结果
    res := make([][]byte, 0, len(fields))
    var resErr error
    //2.将多个key分批获取（每次32个）
    pageNum := int(math.Ceil(float64(len(fields)) / float64(_CHUNK_SIZE)))
    for i := 0; i < pageNum; i++ {
        //2.1创建分批切片 []string
        var end int
        if i == (pageNum - 1) {
            end = len(fields)
        } else {
            end = (i + 1) * _CHUNK_SIZE
        }
        chunk := fields[i*_CHUNK_SIZE : end]
        //2.2分批切片的类型转换 => [][]byte
        chunkLength := len(chunk)
        fieldList := make([]interface{}, 0, chunkLength)
        for _, v := range chunk {
            fieldList = append(fieldList, v)
        }
        cacheRes, err := libRedis.ByteSlices(entity.Do("HMGET", libRedis.Args{}.Add(key).AddFlat(fieldList)...))
        if err != nil {
            for i := 0; i < chunkLength; i++ {
                res = append(res, nil)
            }
            entity.LogWarn("cache_mget_error: ", err)
            continue
        } else {
            res = append(res, cacheRes...)
        }
    }
    return res, resErr
}

// HMSet 将一个map存到Redis hash
func (entity *Redis) HMSet(key string, fvmap map[string]interface{}) error {
    _, err := entity.Do("HMSET", libRedis.Args{}.Add(key).AddFlat(fvmap)...)
    return err
}

func (entity *Redis) HKeys(key string) ([][]byte, error) {
    if res, err := libRedis.ByteSlices(entity.Do("HKEYS", key)); err == libRedis.ErrNil {
        return nil, nil
    } else {
        return res, err
    }
}

func (entity *Redis) HGetAll(key string) ([][]byte, error) {
    if res, err := libRedis.ByteSlices(entity.Do("HGETALL", key)); err == libRedis.ErrNil {
        return nil, nil
    } else {
        return res, err
    }
}

func (entity *Redis) HLen(key string) (int64, error) {
    if res, err := libRedis.Int64(entity.Do("HLEN", key)); err == libRedis.ErrNil {
        return 0, nil
    } else {
        return res, err
    }
}

func (entity *Redis) HVals(key string) ([][]byte, error) {
    if res, err := libRedis.ByteSlices(entity.Do("HVALS", key)); err == libRedis.ErrNil {
        return nil, nil
    } else {
        return res, err
    }
}

func (entity *Redis) HIncrBy(key, field string, value int64) (int64, error) {
    return libRedis.Int64(entity.Do("HINCRBY", key, field, value))
}

func (entity *Redis) HExists(key string, field string) (bool, error) {
    if res, err := libRedis.Bool(entity.Do("HEXISTS", key, field)); err == libRedis.ErrNil {
        return false, nil
    } else {
        return res, err
    }
}

func (entity *Redis) Get(key string) ([]byte, error) {
    if res, err := libRedis.Bytes(entity.Do("GET", key)); err == libRedis.ErrNil {
        return nil, nil
    } else {
        return res, err
    }
}

func (entity *Redis) MGet(keys ...string) [][]byte {
    //1.初始化返回结果
    res := make([][]byte, 0, len(keys))

    //2.将多个key分批获取（每次32个）
    pageNum := int(math.Ceil(float64(len(keys)) / float64(_CHUNK_SIZE)))
    for n := 0; n < pageNum; n++ {
        //2.1创建分批切片 []string
        var end int
        if n != (pageNum - 1) {
            end = (n + 1) * _CHUNK_SIZE
        } else {
            end = len(keys)
        }
        chunk := keys[n*_CHUNK_SIZE : end]
        //2.2分批切片的类型转换 => []interface{}
        chunkLength := len(chunk)
        keyList := make([]interface{}, 0, chunkLength)
        for _, v := range chunk {
            keyList = append(keyList, v)
        }
        cacheRes, err := libRedis.ByteSlices(entity.Do("MGET", keyList...))
        if err != nil {
            for i := 0; i < len(keyList); i++ {
                res = append(res, nil)
            }
        } else {
            res = append(res, cacheRes...)
        }
    }
    return res
}

func (entity *Redis) MSet(values ...interface{}) error {
    _, err := entity.Do("MSET", values...)
    return err
}

func (entity *Redis) Set(key string, value interface{}, expire ...int64) error {
    var res string
    var err error
    if expire == nil {
        res, err = libRedis.String(entity.Do("SET", key, value))
    } else {
        res, err = libRedis.String(entity.Do("SET", key, value, "EX", expire[0]))
    }
    if err != nil {
        return err
    } else if strings.ToLower(res) != "ok" {
        return errors.New("set result not OK")
    }
    return nil
}

func (entity *Redis) SetEx(key string, value interface{}, expire int64) error {
    return entity.Set(key, value, expire)
}

func (entity *Redis) Append(key string, value interface{}) (int, error) {
    return libRedis.Int(entity.Do("APPEND", key, value))
}

func (entity *Redis) Incr(key string) (int64, error) {
    return libRedis.Int64(entity.Do("INCR", key))
}

func (entity *Redis) IncrBy(key string, value int64) (int64, error) {
    return libRedis.Int64(entity.Do("INCRBY", key, value))
}

func (entity *Redis) IncrByFloat(key string, value float64) (float64, error) {
    return libRedis.Float64(entity.Do("INCRBYFLOAT", key, value))
}

func (entity *Redis) Decr(key string) (int64, error) {
    return libRedis.Int64(entity.Do("DECR", key))
}

func (entity *Redis) DecrBy(key string, value int64) (int64, error) {
    return libRedis.Int64(entity.Do("DECRBY", key, value))
}
