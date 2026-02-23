package controller

type Controller struct {
	User
	Form
	Admin
}

func New() *Controller {
	Controller := &Controller{}
	return Controller
}
