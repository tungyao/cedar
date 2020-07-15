Router on prefix tree algorithm 😀  
---
[session component for cedar](https://github.com/tungyao/cedar-session)
---
# all structure
**cedar.NewRouter().Get(prefix,http.HandlerFunc,http.Handler)**
> Only one can take effect
## update
* Mature MVC Framework
> view->app->index.html

Page is required at the beginning of a method
```go
func PageAppIndex(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
    r.View().Assign("name","hello").Render()
    // or
    r.View().Render("app/index")
}
func AppIndex(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
    r.Json().Success(map[string]string{"name":"cedar"})
    // or
    r.Json().Success(struct{})
}
func TestParam(t *testing.T) {
	r := cedar.NewRouter()
	r.SetDebug()
	r.SetLayout()
	r.Get("/", PageAppIndex, nil)
	http.ListenAndServe(":8000", r)
}

```

* new route rule
```go
r.Get("/index/:id",func(w http.ResponseWriter, r *http.Request){
    fmt.Println(r.URL.Fragment) // <- use this ,get the `id`
},nil)
```
=======
* Middlewre
***You must declare it in advance***

##### `return false is not continue` `return true can be`
```go
r.Middleware("test", func(w http.ResponseWriter, r *http.Request) bool {
	http.Redirect(w, r, "/a/b/c", 302)
	return false
})
r.Get("/", func(writer http.ResponseWriter, request *http.Request,co *Core) {
	writer.Write([]byte("hello"))
}, nil, "test") <- middleware name
```
* Add dynamic route ,use yaml file to generate route ,must be like this
```yaml
- path: /dynamic/day
  view: /static/dynamic1.html
- path: /dynamic/day2
  view: /static/dynamic2.html
```
* Add new function: global function
>  it can to record logs and so on
```go
r.GlobalFunc("test", func(w http.ResponseWriter, r *http.Request,co *Core) error {
	fmt.Println("global func run")
	return nil
})
````
# Example
Normal
```
r := cedar.NewRouter()
r.Get("/",HandlerFunc(),nil)
r.Post("/",HandlerFunc(),nil)
r.Put("/",HandlerFunc(),nil)
r.Delete("/",HandlerFunc(),nil)
if err := http.ListenAndServe(":80", r); err != nil {
	log.Panicln(err)
}
```
Group
```
r := cedar.NewRouter()
r.Group("/",func (group *cedar.Groups){
    group.Get("/",HandlerFunc(),nil)
    group.Group("/x",func(groups *cedar.Groups) {
        group.Get("/x",HandlerFunc(),nil)
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
