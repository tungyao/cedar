package router

import (
	"net/http"

	"../../../cedar"
)

type Auto cedar.AutoRegister

func (a *Auto) GetPage(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
	writer.Write([]byte("hello"))
}
