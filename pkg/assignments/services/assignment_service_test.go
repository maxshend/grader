package services

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/maxshend/grader/pkg/assignments"
)

func TestAssignmentsGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := assignments.NewMockRepositoryInterface(ctrl)
	service := NewAssignmentsService("", repo, nil, nil, nil, "", "")
	var id int64 = 1

	t.Run("success", func(t *testing.T) {
		assignment := &assignments.Assignment{ID: id}
		repo.EXPECT().GetByID(id).Return(assignment, nil)

		result, err := service.GetByID(id)
		if err != nil {
			t.Fatalf("expected to not have errors got %v", err)
		}
		if !reflect.DeepEqual(result, assignment) {
			t.Errorf("expect to have %+v, got %+v", assignment, result)
		}
	})

	t.Run("error", func(t *testing.T) {
		repo.EXPECT().GetByID(id).Return(nil, fmt.Errorf("db_error"))
		_, err := service.GetByID(id)
		if err == nil {
			t.Fatalf("expected to have errors")
		}
	})
}
