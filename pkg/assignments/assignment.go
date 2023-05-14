package assignments

type Assignment struct {
	ID          int64
	Title       string
	Description string
	GraderURL   string
	Container   string
	PartID      string
	Files       []string
}

type RepositoryInterface interface {
	GetAll(limit int, offset int) ([]*Assignment, error)
}
