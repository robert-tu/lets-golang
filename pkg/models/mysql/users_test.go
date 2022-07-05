package mysql

import (
	"reflect"
	"testing"
	"time"

	"robert-tu.net/snippetbox/pkg/models"
)

func TestUserModelGet(t *testing.T) {
	// skip test if -short flag
	if testing.Short() {
		t.Skip("mysql: skipping integration test")
	}

	// table-driven tests
	tests := []struct {
		name      string
		userID    int
		wantUser  *models.User
		wantError error
	}{
		{
			name:   "Valid",
			userID: 1,
			wantUser: &models.User{
				ID:      1,
				Name:    "Bob Jones",
				Email:   "bob@gmail.com",
				Created: time.Date(2020, 12, 31, 11, 0, 0, 0, time.UTC),
				Active:  true,
			},
			wantError: nil,
		},
		{
			name:      "Zero ID",
			userID:    0,
			wantUser:  nil,
			wantError: models.ErrNoRecord,
		},
		{
			name:      "Non-existent ID",
			userID:    2,
			wantUser:  nil,
			wantError: models.ErrNoRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialize connection pool
			db, teardown := newTestDB(t)
			defer teardown()

			// create new instance
			m := UserModel{db}

			// call Get() method
			user, err := m.Get(tt.userID)

			if err != tt.wantError {
				t.Errorf("want %v; got %s", tt.wantError, err)
			}

			// checks complex custom types
			if !reflect.DeepEqual(user, tt.wantUser) {
				t.Errorf("want %v; got %v", tt.wantUser, user)
			}
		})
	}
}
