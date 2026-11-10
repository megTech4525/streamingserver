package user

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// func (s *Service) RegisterUser(ctx context.Context, username, email, password string) (*User, error) {
// 	//
// }
