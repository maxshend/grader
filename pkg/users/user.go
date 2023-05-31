package users

const (
	DefaultProvider int = iota
	VkProvider
)

type User struct {
	ID       int64
	Username string
	Password string
	Provider int
	IsAdmin  bool
}

type RepositoryInterface interface {
	GetAll(limit int, offset int) ([]*User, error)
	Create(username, password string, provider int, isAdmin bool) (*User, error)
	GetByID(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByUsernameProvider(username string, provider int) (*User, error)
	Update(*User) (*User, error)
}
