package users

const (
	RegularUser int = iota
	Admin
)

type User struct {
	ID       int64
	Username string
	Password string
	IsAdmin  bool
	Role     int
}

type RepositoryInterface interface {
	Create(username, password string, role int) (*User, error)
	GetByID(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
}
