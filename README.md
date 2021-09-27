CtxFlow是一个轻薄的业务分层框架。

### 2021.9.27变更
* Dao层提供了SetModel、GetModel方法，可用于替换原来的SetTable

### 2021.7.14变更
* 提供了统一的初始化方法puzzle.InitConfig(请参考examples)
* 初始化时支持忽略DbLog的默认格式化方式(IgnoreDefaultDBLogFormat)，原因是有些框架会在gorm底层提供默认的格式化方式

### 2021.7.13变更
* 调整了api模块中使用http客户端的方式，由直接将方法封装到api基类变成了通过adapter进行封装。使用者可以根据适配器demo实现更丰富更灵活的功能
* 新增demo_adapter文件(里面的代码注释掉了)，用于给使用者做参考
* tag版本由2.x变为1.10.x(因为2.x使用引用路径比较麻烦，后续将基于1.10.x向后升级，2.x版本tag仍然可用)
* puzzle.IHttpClient不再需要

### 2021.5.20变更
* 为了便于理解原Domain模块，更名为DataSet
* api层依赖的http工具放入适配器中
* example中增加adapter，用于放置适配器

### V1.10版本重要变更
* CtxFlow从v1.10.x开始,全面支持gormV2,不再支持gormV1
* dao层log简化，直接使用Flow基类提供的log输出mysql的基本信息

### 主要功能
* 结合其他主流框架快速搭建项目
* 兼容主流的gin、gorm、redigo、zap等类库
* 提供全局的上下文
* 全局使用log以及log追踪
* 面向对象的golang代码编写风格
* 支持跨模块事务操作
 
### 框架并不关心的内容(大部分团队是有能力做到)
* gin提供的并发以及接受请求的封装
* 统一的配置加载方式和配置文件相关功能
* 平滑重启以及部署相关逻辑
* log文件以及log对象的初始配置
* 数据库对象的初始配置
* redis对象的初始化配置

### DEMO使用
* examples目录包含几个运行demo,可以copy出来运行
* 搭建demo需要在本地搭建mysql数据库，并且main.go中设置账号密码
* 数据库搭建好后需要创建demo需要的表
```sql
    create database demo;
    use demo;
    CREATE TABLE `demoUser` (
      `uid` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
      `name` varchar(512)  NOT NULL DEFAULT '' COMMENT '用户名',
      `age` int(20) unsigned NOT NULL DEFAULT '0' COMMENT '年龄',
      `last_modify_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
      PRIMARY KEY (`uid`)
    ) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8 COMMENT='用户demo';
    insert into demoUser (`name`,`age`) values ("张三","20");
````
* demo访问方法 需要安装jq(用来格式化JSON)命令
```shell script
curl -X POST 'http://localhost:8989/demo/testLog' | jq
curl -X POST 'http://localhost:8989/demo/testGetUserList' | jq
curl -X POST 'http://localhost:8989/demo/testAddUser' --data '{"name":"李四","age":11}' | jq
curl -X POST 'http://localhost:8989/demo/testHttpGet' | jq
````
* 可以通过examples/log.txt查看log

### 业务分层建议
通常业务代码建议按照业务层次划分主要分为Controller、Service、Data、Dao、Api等多层应用架构。

### Controller(控制器)
Controller层是调度层，主要的职责是：

* 接收接口入参
* 参数校验
* 调度Service完成服务
* 输出数据结果

#### 控制器执行原理
Controller层代码核心是一个继承自layer.Controller的结构体，并且可以使用继承自layer.Controller提供的一系列方法。
```go
type TestAddUser struct {
    layer.Controller
}

func (entity *TestAddUser) Action() {
    entity.LogInfo("test getUser start")

    req := dtoUser.AddUserReq{
        Age: thrift.Int32Ptr(10),//给age赋予默认值
    }

    if err :=entity.BindParamError(&req);err != nil{
        entity.RenderJsonFail(err)
        return
    }

    userService := entity.Use(new(svUser.UserService)).(*svUser.UserService)
    err := userService.AddUser(&req)
    if err != nil {
        entity.RenderJsonFail(err)
        return
    }
    entity.RenderJsonSucc("success")
}
```
Controller中结构体的入口方法方法是Action方法。路由层入口配置如下：

```go
    demoGroup := engine.Group("/demo")
    {
        demoGroup.POST("/testLog", ctxflow.UseController(new(demo.TestLog)))
        demoGroup.POST("/testGetUserList", ctxflow.UseController(new(demo.TestGetUserList)))
        demoGroup.POST("/testAddUser", ctxflow.UseController(new(demo.TestAddUser)))
        demoGroup.POST("/testHttpGet", ctxflow.UseController(new(demo.TestHttpGet)))
    }
    engine.Run("0.0.0.0:8989")
```

其中与正常入口配置的方式有所区别。

正常的入口配置是POST方法第二个参数传入一个函数指针。

而ctxflow则是使用ctxflow.UseController(new(data.GetData))的方式。

    该方式表示默认初始化的是以Controller的实例化对象为调用目标。
    初始化后，会调用Action方法。同时会将上下文信息写入到控制器类中。
控制器提供的方法：       
* BindParam(req interface{}) bool
* Action()
* RenderJsonFail(err error)
* RenderJsonSucc(data interface{}) 
* RenderJsonAbort(err error)
* 继承自CtxFlow的其他方法


### Controller编码规范建议
* router入口必须使用ctxflow.UseController的方式进入
* 需要定义继承于layer.Controller的结构体，并实现Action方法
* 入参的Dto结构体定义在dto目录，结构体遵循validator v9的规范和写法https://godoc.org/gopkg.in/go-playground/validator.v9。结构体中可选参数尽量使用指针。
* 入参的结构体需要赋予默认值时，参考上面范例代码的声明方式。
* 内容数据输出使用layer.Controller基类提供的RenderXXX方法输出
* 结构体内对象本身的引用命名统一命名成entity
* 使用其他模块时，要使用如下方式进行初始化对象，注意后面要有指针类型的强制转换
userService := entity.Use(new(svUser.UserService)).(*svUser.UserService)
注：Use方法会将上下文信息传导到其他的模块中
* 可以通过entity.Use的第二个（或以上）参数扩展传递，并且这些参数在被use的模块中是可继承的


### Service(服务层)
Service层的的代码是业务实现的主要模块，包括

#### 主逻辑控制
* 调用data层获取数据，进行拼装业务逻辑
* 对最终的结果负责

#### Service层的主要执行原理
* Service层需要使用entity.Use方法进行初始化。调用其他模块也需要通过entity.Use进行初始化替他模块。

### Service层编码规范建议
* 需要使用entity.Use进行初始化，和初始化使用的其他模块
* 主要的业务实现逻辑放在Service中
* 需要以结构体对象的方式声明Service,Service需要继承于layer.Service
* service层的包名以sv前缀开头
* req结构体只可以在这一层和 Controller层出现
* 可以通过entity.Use的第二个（或以上）参数扩展传递，并且这些参数在被use的模块中是可继承的

### Data
Data层的代码是针对对象个体实现的层，主要功能是

* 能够抽象的最细颗粒度对象的实现放在Data层


### Data层编码规范建议

* 需要使用entity.Use进行初始化，和初始化使用的其他模块
* 最细粒度对象的实现放在Data层
* 需要以结构体对象的方式声明Data,Data需要继承于layer.DataSet
* Data层的包名以ds前缀开头
* req结构体不可以传入到该层
* 可以通过entity.Use的第二个（或以上）参数扩展传递，并且这些参数在被use的模块中是可继承的

### Api(接口层)
Api层主要是对与其他服务ral/rpc调用的封装

### Api层编码规范建议

* 需要使用entity.Use进行初始化，和初始化使用的其他模块
* Api调用的基本封装
* 需要以结构体对象的方式声明Api,Api需要继承于BaseApi
* api层的包名以api前缀开头
* 使用BaseApi提供的方法进行http调用
* preUse方法相当于Use的前置方法，可以当成钩子(构造)函数使用，需要在preUse方法中初始化调用api的配置
* 可以通过entity.Use的第二个（或以上）参数扩展传递，并且这些参数在被use的模块中是可继承的

### Dao(数据库层)
dao层主要是对数据库操作的封装，与Api层封装相对应

### Dao层编码规范建议

* 需要使用entity.Use进行初始化（其他模块也是如此）
* 本模块是数据库操作的基本封装
* 需要以结构体对象的方式声明dao,dao需要继承于BaseDao
* preUse方法相当于Use的前置方法，可以当成钩子(构造)函数使用，需要在preUse方法中初始化调用数据表的配置
* dao层的包名以dao前缀开头
* 全局事务的使用，使用全局事务可以在逻辑中先声明好db对象，如
db := puzzle.GetDefaultGormDb().Begin()
* 然后entity.Use模块的时候将该db传入（第二个参数）。Use的模块可以是任何模块（包括Service,Data,Dao等）。当这个模块里使用了Dao层的时候，会自动使用这个db，然后逻辑里就可以使用这个db进行统一的提交和回滚操作了。

db使用的优先级--假如依赖关系是A(Service)->B(Data)->C(Dao),既A中通过Use使用了B,B中通过Use使用了C。那么最后C执行的时候会是这样：
1.如果只有在A  use B的时候传入了db1,则最终C中执行的是db1这个gorm.DB对象
2.如果步骤1中，A  use B的时候传入了db1,同时B use C中传入了db2,则最终C中执行的是db2这个gorm.DB对象，遵循最近的db最优先的原则。

示例代码：
```go
    //db关联其他模块，统一提交事务或者回滚
    db := puzzle.GetDefaultGormDb().Begin()
    //use方法的第二个参数可选，可以是db也可以是其他。如果设置为某个DB，则被这个DB关联了事务
    userData := entity.Use(new(dsUser.UserRepository),db).(*dsUser.UserRepository)
    usr,err:=userData.GetUserByName(req.Name)

    if err != gorm.ErrRecordNotFound {
        entity.LogWarn("用户已存在:%+v",usr)
        db.Rollback()
        return err
    }

    //use方法的第二个参数可选，可以是db也可以是其他。如果设置为某个DB，则被这个DB关联了事务
    userDao := entity.Use(new(daoUser.DemoUserDao),db).(*daoUser.DemoUserDao)
    err = userDao.Create(&daoUser.DemoUser{
        Name: req.Name,
        Age:*req.Age,
    })

    if err != nil {
        entity.LogWarn("创建失败:%+v",usr)
        db.Rollback()
        return err
    }
    db.Commit()
```
注：当对象被Use的过程中，Use的第2个以上的扩展参数会自动继承到这个对象所有Use的其他对象中，可以通过entity.GetArgs(idx)获取。如果第一个参数是*gorm.DB类型，则里面所有使用的Dao模块都会使用这个DB（就近原则）。


## CtxFlow(上下文流)
所有模块的基类都要继承于CtxFlow类，该类提供了Log相关的方法以及Use、SetContext、GetContext、PreUse等方法。各个模块都可以使用

