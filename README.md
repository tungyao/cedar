Router on prefix tree lookup algorithm 😀  
---
# all structure
**cedar.NewRouter().Get(prefix,http.HandlerFunc,http.Handler)**
> Only one can take effect
# Example
Normal
```
r := cedar.NewRouter()
r.Get("/",http.HandlerFunc(),nil)
r.Post("/",http.HandlerFunc(),nil)
r.Put("/",http.HandlerFunc(),nil)
r.Delete("/",http.HandlerFunc(),nil)
```
Group
```
r := cedar.NewRouter()
r.Group("/",func (group *cedar.Groups){
    group.Get("/",http.HandlerFunc())
})
```
---
RestFul 
```go
r := cedar.NewRestRouter(cedar.RestConfig{
		EntryPath: "yashua",
		ApiName:   "api",
        Pattern:"." `new*`

})
r.GetR(api,fn)
r.PostR(api,fn)
r.PutR(api,fn)
r.DeleteR(api,fn)
r.GroupR(path,func(groups *cedar.GroupR{
    r.GetR(api,fn)
})
```
# exp
```
r.GetR("user.add", func(writer http.ResponseWriter, request *http.Request) {
 		_, _ = fmt.Fprintln(writer, "hello")
})
```
`localhost/wechat?api=user.add`  *The "Pattern" is there ,you can use other  punctuation marks*

[Other Exp](https://github.com/tungyao/cedar/blob/master/test/route_test.go)

### Usage
next time
