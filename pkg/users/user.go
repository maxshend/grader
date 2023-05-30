package users

const (
	RegularUser int = iota
	Admin
)

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
	Role     int
}

type RepositoryInterface interface {
	Create(username, password string, provider, role int) (*User, error)
	GetByID(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByUsernameProvider(username string, provider int) (*User, error)
}
