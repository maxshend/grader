package utils

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func CheckSqlMockExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
