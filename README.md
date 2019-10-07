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
r.Static(static dir path)
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
})
r.GetR(api,fn)
r.PostR(api,fn)
r.PutR(api,fn)
r.DeleteR(api,fn)

entry url => localhost/wechat?api=user
```
### Usage
next time
