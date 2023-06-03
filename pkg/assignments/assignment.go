package assignments

type Assignment struct {
	ID          int64
	CreatorID   int64
	Title       string
	Description string
	GraderURL   string
	Container   string
	PartID      string
	Files       []string
}

type RepositoryInterface interface {
	GetAllByCreator(creatorID int64, limit int, offset int) ([]*Assignment, error)
	GetByID(int64) (*Assignment, error)
	GetByIDByCreator(id int64, creatorID int64) (*Assignment, error)
	GetByTitle(string) (*Assignment, error)
	GetByUserID(userID int64, limit, offset int) ([]*Assignment, error)
	Create(creatorID int64, title, description, graderURL, container, partID string, files []string) (*Assignment, error)
	Update(*Assignment) (*Assignment, error)
}
