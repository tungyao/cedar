package router

import (
	"net/http"

	"../../../cedar"
)

type Auto cedar.AutoRegister

func (a *Auto) GetPageAppIndex(writer http.ResponseWriter, request *http.Request, r *cedar.Core) {
	writer.Write([]byte("hello"))
}
func (a *Auto) GetMTestAppIndex() {

}
