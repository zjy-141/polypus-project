package service

type Service struct {
	Admin
	Form
	User
}

func New() *Service {
	service := &Service{}
	return service
}
