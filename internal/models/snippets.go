package models

import (
	"database/sql"
	"errors"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (model *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	statement := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := model.DB.Exec(statement, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (model *SnippetModel) Get(id int) (Snippet, error) {
	statement := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`

	row := model.DB.QueryRow(statement, id)

	var snip Snippet

	err := row.Scan(&snip.ID, &snip.Title, &snip.Content, &snip.Created, &snip.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return snip, nil
}

func (model *SnippetModel) Latest() ([]Snippet, error) {
	statement := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	rows, err := model.DB.Query(statement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var snippets []Snippet

	for rows.Next() {
		var snip Snippet

		err = rows.Scan(&snip.ID, &snip.Title, &snip.Content, &snip.Created, &snip.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, snip)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
