package attachments

import "io"

type Attachment struct {
	URL     string
	Content io.Reader
	Name    string
}

type RepositoryInterface interface {
	Create(pathPrefix, name string, content io.Reader) (*Attachment, error)
	Destroy(path string) error
}
