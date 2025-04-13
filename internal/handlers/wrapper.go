package handlers

import "github.com/wisp167/pvz/api"

func RegisterHandlersMiddleware(router api.EchoRouter, si api.ServerInterface) {
	RegisterHandlersMiddlewareWithBaseURL(router, si, "")
}

// i want to apply auth middleware before passing context to wrapper (thus redefinition)
func RegisterHandlersMiddlewareWithBaseURL(router api.EchoRouter, si api.ServerInterface, baseURL string) {

	wrapper := api.ServerInterfaceWrapper{
		Handler: si,
	}

	//moderatorOnly := RoleRequired("moderator")
	employeeOnly := RoleRequired("employee")
	moderatorOnly := RoleRequired("moderator")

	router.POST(baseURL+"/dummyLogin", wrapper.PostDummyLogin)
	router.POST(baseURL+"/login", wrapper.PostLogin)
	router.POST(baseURL+"/products", wrapper.PostProducts, employeeOnly)
	router.GET(baseURL+"/pvz", wrapper.GetPvz)
	router.POST(baseURL+"/pvz", wrapper.PostPvz, moderatorOnly)
	router.POST(baseURL+"/pvz/:pvzId/close_last_reception", wrapper.PostPvzPvzIdCloseLastReception, employeeOnly)
	router.POST(baseURL+"/pvz/:pvzId/delete_last_product", wrapper.PostPvzPvzIdDeleteLastProduct, employeeOnly)
	router.POST(baseURL+"/receptions", wrapper.PostReceptions, employeeOnly)
	router.POST(baseURL+"/register", wrapper.PostRegister)

}
