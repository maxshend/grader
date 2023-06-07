package repo

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/maxshend/grader/pkg/attachments"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmpDir := t.TempDir()
	host := "http://localhost"
	repo := NewAttachmentsInmemRepo(host, tmpDir)

	rawData := []byte("Hello")
	prefix := "test_dir"
	filename := "test.txt"
	url := filepath.Join(tmpDir, prefix, filename)

	content := bytes.NewReader(rawData)
	expected := &attachments.Attachment{
		Name:    filename,
		URL:     host + url,
		Content: content,
	}
	attachment, err := repo.Create(prefix, filename, content)

	if err != nil {
		t.Fatalf("expected not to have errors, got %v", err)
	}
	if !reflect.DeepEqual(attachment, expected) {
		t.Fatalf("expected %+v got %+v", expected, attachment)
	}

	writtenData, _ := os.ReadFile(url)
	if !reflect.DeepEqual(writtenData, rawData) {
		t.Fatalf("expected %v got %v", rawData, writtenData)
	}
}

func TestDestroy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmpDir := t.TempDir()
	repo := NewAttachmentsInmemRepo("/", tmpDir)

	t.Run("success", func(t *testing.T) {
		rawData := []byte("Hello")
		prefix := "test_dir"
		filename := "test.txt"
		content := bytes.NewReader(rawData)

		attachment, err := repo.Create(prefix, filename, content)
		if err != nil {
			t.Fatalf("error creating test file %v", err)
		}

		err = repo.Destroy(attachment.URL)
		if err != nil {
			t.Fatalf("expected not to have errors, got %v", err)
		}
		_, err = os.Stat(attachment.URL)
		if !os.IsNotExist(err) {
			t.Fatalf("expected to remove attachment %+v", attachment)
		}
	})

	t.Run("error", func(t *testing.T) {
		err := repo.Destroy(".invalid_path")
		if err == nil {
			t.Fatalf("expected to have errors")
		}
	})
}
