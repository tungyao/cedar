package v1

import (
	"net/http"

	"../../../../cedar"
)

type M2 cedar.AutoRegister

func (a *M2) Get_test_PageM2Index(writer http.ResponseWriter, request *http.Request, co *cedar.Core) {
	co.View().Assign("name", "cedar").Render()
}
