package repo

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/maxshend/grader/pkg/assignments"
	"github.com/maxshend/grader/pkg/utils"
)

func TestAssignmentsGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAssignmentsSQLRepo(db)
	sqlQuery := "SELECT id, title, description"
	fields := []string{"id", "title", "description", "grader_url", "container", "part_id", "files"}
	var assignmentID int64 = 1

	type testCase struct {
		Title   string
		Success bool
		Want    *assignments.Assignment
		Mock    func(*testing.T, *testCase, *sqlmock.ExpectedQuery)
		Check   func(*testing.T, *testCase, *assignments.Assignment)
	}

	testCases := []*testCase{
		{
			Title:   "Success",
			Success: true,
			Want:    &assignments.Assignment{ID: assignmentID},
			Mock: func(t *testing.T, tc *testCase, expected *sqlmock.ExpectedQuery) {
				t.Helper()

				tc.Want.Files = []string{"main.go"}
				files := "{\"main.go\"}"
				rows := sqlmock.NewRows(fields).AddRow(
					tc.Want.ID, tc.Want.Title, tc.Want.Description, tc.Want.GraderURL,
					tc.Want.Container, tc.Want.PartID, files,
				)

				expected.WithArgs(tc.Want.ID).WillReturnRows(rows)
			},
			Check: func(t *testing.T, tc *testCase, got *assignments.Assignment) {
				if !reflect.DeepEqual(got, tc.Want) {
					t.Fatalf("expected %+v, got %+v", tc.Want, got)
				}
			},
		},
		{
			Title:   "Not found",
			Success: true,
			Want:    &assignments.Assignment{ID: assignmentID},
			Mock: func(t *testing.T, tc *testCase, expected *sqlmock.ExpectedQuery) {
				rows := sqlmock.NewRows(fields)

				expected.WithArgs(tc.Want.ID).WillReturnRows(rows)
			},
			Check: func(t *testing.T, tc *testCase, got *assignments.Assignment) {
				if got != nil {
					t.Fatalf("expected result to be nil, got %+v", got)
				}
			},
		},
		{
			Title:   "DB Error",
			Success: false,
			Want:    &assignments.Assignment{ID: assignmentID},
			Mock: func(t *testing.T, tc *testCase, expected *sqlmock.ExpectedQuery) {
				expected.WithArgs(tc.Want.ID).WillReturnError(fmt.Errorf("db_error"))
			},
			Check: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			defer utils.CheckSqlMockExpectations(t, mock)

			testCase.Mock(t, testCase, mock.ExpectQuery(sqlQuery))

			got, err := repo.GetByID(testCase.Want.ID)
			if testCase.Success {
				if err != nil {
					t.Fatalf("expect not to have errors, got %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expect to have errors")
				}
			}

			if testCase.Check != nil {
				testCase.Check(t, testCase, got)
			}
		})
	}
}
