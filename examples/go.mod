module examples

go 1.13

//如果独立运行，这一行是需要注释掉的（不注释会直接使用本地库）
replace github.com/go-opener/ctxflow => ../../ctxflow

require (
	github.com/apache/thrift v0.12.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-opener/ctxflow v1.10.2
	github.com/json-iterator/go v1.1.10
	go.uber.org/zap v1.16.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gorm.io/driver/mysql v1.0.6
	gorm.io/gorm v1.21.9
)
