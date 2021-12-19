package ultimate_cedar

type MiddlewareInterceptor func(ResponseWriter, Request, HandlerFunc)
type MiddlewareChain []MiddlewareInterceptor
type MiddlewareHandlerFunc HandlerFunc

func (cont MiddlewareHandlerFunc) Intercept(mw MiddlewareInterceptor) MiddlewareHandlerFunc {
	return func(writer ResponseWriter, request Request) {
		mw(writer, request, HandlerFunc(cont))
	}
}
func (chain MiddlewareChain) Handler(handlerFunc HandlerFunc) HandlerFunc {
	curr := MiddlewareHandlerFunc(handlerFunc)
	for i := len(chain) - 1; i >= 0; i-- {
		curr = curr.Intercept(chain[i])
	}
	return HandlerFunc(curr)
}
