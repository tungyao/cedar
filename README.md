Router on prefix tree algorithm 😀  
---
[session component for cedar](https://github.com/tungyao/cedar-session)
---
# all structure
**cedar.NewRouter().Get(prefix,http.HandlerFunc,http.Handler)**
> Only one can take effect
## update
* Add new function: global function
>  it can to record logs and so on
 ```go
r.GlobalFunc("test", func(w http.ResponseWriter,r *http.Request) error {
		fmt.Println("123213")
		return nil
})
```
# Example
Normal
```
r := cedar.NewRouter()
r.Get("/",http.HandlerFunc(),nil)
r.Post("/",http.HandlerFunc(),nil)
r.Put("/",http.HandlerFunc(),nil)
r.Delete("/",http.HandlerFunc(),nil)
if err := http.ListenAndServe(":80", r); err != nil {
	log.Panicln(err)
}
```
Group
```
r := cedar.NewRouter()
r.Group("/",func (group *cedar.Groups){
    group.Get("/",http.HandlerFunc(),nil)
    group.Group("/x",func(groups *cedar.Groups) {
        group.Get("/x",http.HandlerFunc(),nil)
    })
})
if err := http.ListenAndServe(":80", r); err != nil {
	log.Panicln(err)
}
```
---
RestFul 
```go
r := cedar.NewRestRouter(cedar.RestConfig{
		EntryPath: "yashua",
		ApiName:   "api",
        Pattern:"." `new*`

})
r.Get(api,fn,handler)
r.Post(api,fn,handler)
r.Put(api,fn,handler)
r.Delete(api,fn,handler)
r.Group(path,func(groups *cedar.Group{
    r.Get(api,fn,handler)
})
```
# exp
```
r.Get("user.add", func(writer http.ResponseWriter, request *http.Request) {
 		_, _ = fmt.Fprintln(writer, "hello")
})
```
`localhost/wechat?api=user.add`  *The "Pattern" is there ,you can use other  punctuation marks*

[Other Exp](https://github.com/tungyao/cedar/blob/master/test/route_test.go)

### Usage
next time
