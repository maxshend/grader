package repo

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/maxshend/grader/pkg/attachments"
)

type AttachmentsInmemRepo struct {
	Path string
	mx   sync.Mutex
}

func NewAttachmentsInmemRepo(path string) *AttachmentsInmemRepo {
	return &AttachmentsInmemRepo{Path: path}
}

func (r *AttachmentsInmemRepo) Create(pathPrefix, name string, content io.Reader) (*attachments.Attachment, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	fullpath := filepath.Join(r.Path, pathPrefix, name)
	if err := os.MkdirAll(filepath.Dir(fullpath), 0755); err != nil {
		return nil, err
	}

	newFile, err := os.Create(fullpath)
	if err != nil {
		return nil, err
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, content); err != nil {
		return nil, err
	}

	attachment := &attachments.Attachment{Name: name, Content: content, URL: fullpath}

	return attachment, nil
}
