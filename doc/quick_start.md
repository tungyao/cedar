> 推荐使用 `go.mod`

- 引入最新版api
  
`import "github.com/tungyao/cedar"`
- 声明一个路由， `NewRouter` 携带连个参数，只是用来区分session，非必填
    - 第一个：`self` ，当前网站的域名，比如`self.google.com`
    - 第二个：`domino` ，顶级域名，比如`google.com`，用来跨域名方便
  
```go
r:=cedar.NewRouter()
r.StopSession() // 将不在设置cookie
r.SetDebug()    // 将错误信息显示
```

- 书写路由
    - Get
    - Post
    - Head
    - Put
    - Patch
    - Delete
    - Connect
    - Trace
    - Options
    
```go
r.Get(path string, handlerFunc HandlerFunc, middleName ...string)
```

- 中间件加载方式
```go
r.MiddleWare()
```