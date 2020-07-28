package v1

import (
	"fmt"
	"net/http"

	"../../../../cedar"
)

type Auto cedar.AutoRegister

func (a *Auto) Get_test_PageAppIndex(writer http.ResponseWriter, request *http.Request, co *cedar.Core) {
	co.View().Assign("name", "cedar").Render()
}
func (a *Auto) Get_Index(writer http.ResponseWriter, request *http.Request, co *cedar.Core) {

}
func (a *Auto) MiddleTest(w http.ResponseWriter, r *http.Request, c *cedar.Core) bool {
	fmt.Println("load middleware => test")
	return true
}
