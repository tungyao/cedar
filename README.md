Router on prefix tree lookup algorithm 😀  
---
# Example
Normal
```
r := cedar.NewRouter()
r.Get("/",http.HandlerFunc())
r.Post("/",http.HandlerFunc())
r.Put("/",http.HandlerFunc())
r.Delete("/",http.HandlerFunc())
r.Static("./static/")
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
		EntryPath: "wechat",
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
