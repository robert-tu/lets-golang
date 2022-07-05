package mysql

import (
	"database/sql"
	"errors"

	"robert-tu.net/snippetbox/pkg/models"
)

// define SnippetModel which wraps sql.DB
type SnippetModel struct {
	DB *sql.DB
}

// insert
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires)
    		VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil

}

// get
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires
			FROM snippets
			WHERE expires > UTC_TIMESTAMP() and id = ?`

	// QueryRow() to return pointer of object
	row := m.DB.QueryRow(stmt, id)
	// pointer to new Snippet struct
	s := &models.Snippet{}

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

// top 10
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires
			FROM snippets
			WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	// defer to ensure resultset is always closed before method returns
	// defer comes after Query() to ensure correct resultset
	defer rows.Close()

	// initialize empty slice
	snippets := []*models.Snippet{}
	// iterate
	for rows.Next() {
		// pointer for Snippet struct
		s := &models.Snippet{}
		// copy values from each into Snippet object
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
		// after iteration, check errors with Err()
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}

	return snippets, nil
}
