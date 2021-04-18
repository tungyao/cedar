# 简明
> 这不是一个web框架 ，而是一个路由框架，需要依赖 `http.Handler`，适合一切实现了interface的web框架，包含`net/http`，`http2`等框架
### 在基础路由的方法上，增加了一系列使用功能
- 模糊路由(只能放在最后 ，如果需要放在中间，请贡献你的代码上来)
> `/doc/:id` 
- `Core`结构体，只要在增加core下的方法就能使用
- `cedar.Plugin` 新的插件方法
```go
var Data map[string]string

type TestPlugin struct {
	cedar.Plugin
}

func (tp *TestPlugin) AutoStart() *TestPlugin {
	fmt.Println("插件初始加载")
	Data = make(map[string]string)
	return &TestPlugin{}
}
func (tp *TestPlugin) AutoBefore(w http.ResponseWriter, r *http.Request, co *cedar.Core) {
	fmt.Println("插件运行前加载")
	fmt.Println(Data)
}
func (tp *TestPlugin) Set(key, value string) {
	Data[key] = value
}
func (tp *TestPlugin) Get(key string) string {
	return Data[key]
}
func PageAppIndex(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
	r.Plugin("TestPlugin").Call("Set", request.URL.Query().Get("key"), request.URL.Query().Get("key"))
}
```