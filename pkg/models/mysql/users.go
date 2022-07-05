package mysql

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"robert-tu.net/snippetbox/pkg/models"
)

type UserModel struct {
	DB *sql.DB
}

// Insert
func (m *UserModel) Insert(name, email, password string) error {
	// create bcrypt hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
			VALUES (?, ?, ?, UTC_TIMESTAMP())`

	// use Exec to insert detail
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		// check for MySQLError
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			// checks error related to user_uc_email key
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return models.ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

// Authenticate to verify user exists
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// retrieve id and hashed password
	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password
			FROM users
			WHERE email = ? AND active = TRUE`
	row := m.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		// chceck if user exists
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// check password match
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

// Get
func (m *UserModel) Get(id int) (*models.User, error) {
	u := &models.User{}

	stmt := `SELECT id, name, email, created, active
			FROM users where id = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return u, nil
}
