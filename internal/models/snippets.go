package models

import (
	"database/sql"
	"errors"
	"time"
)

// Same structure as in the snippets table in the database. This struct will be used to hold the data for a single snippet.
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title, content string, expires int) (int, error) {
	stmt := `
	INSERT INTO snippets (title, content, created, expires)
	 VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))
	`

	// Exec is for the statements that don't return rows, like INSERT, UPDATE, DELETE. It returns a sql.Result which contains the number of rows affected and the ID of the last inserted row (if applicable).
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// LastInsertId returns the integer database ID of the last row that was inserted. This is only relevant for statements that insert new rows, like INSERT. If the statement doesn't insert a new row, or if the database doesn't support this feature, it will return an error.
	// LastInsertId and RowsAffected are not supported by all db drivers, for example PostgreSQL doesn't support LastInsertId, so you would need to use RETURNING clause in your SQL statement and QueryRow instead of Exec.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// The returned id has type int64, so we convert it to int type before returning.
	return int(id), nil
}

// Shorter version:
// func (m *SnippetModel) Get(id int) (*Snippet, error) {
// s := &Snippet{}

// err := m.DB.QueryRow("SELECT ...", id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
// if err != nil {
//     if errors.Is(err, sql.ErrNoRows) {
//             return nil, ErrNoRecord
//         } else {
//              return nil, err
//         }
//     }

//	    return s, nil
//	}

func (m *SnippetModel) Get(id int) (*Snippet, error) {
	stmt := `
SELECT id, title, content, created, expires FROM snippets
WHERE expires > UTC_TIMESTAMP() and id = ?
`

	row := m.DB.QueryRow(stmt, id)
	s := &Snippet{}

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// errors.Is() is an idiomatic way starting from Go 1.13. We can also check errors with equality operator. The difference is it's not possible to check original value of original underlying error using the regular equality operator. errors.Is() works by unwrapping errors as necessary before checking for a match.
		if errors.Is(err, sql.ErrNoRows) {
			// The reason we returning ErrNoRecord instead of sql.ErrNoRows is that it helps encapsulate the model completely, so that our application isn't concerned with the underlying datastore or reliant on datastore-specific errors for its behavior. We could easily swap databases without modifying handlers.
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil

}

func (m *SnippetModel) Latest() ([]*Snippet, error) {
	stmt := `
	SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP()
	ORDER BY created DESC
	LIMIT 10
	`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	// Do this to ensure sql.Rows resultset is always properly closed before Latest() method returns. This defer statement should come *after* we check for an error from the Query() method. Otherwise, if Query() returns an error, we'll get a panic trying to close a nil resultset.

	defer rows.Close()

	snippets := []*Snippet{}

	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by the
	// rows.Scan() method. If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		s := &Snippet{}
		err := rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snippets, nil
	// Closing a resultset with defer rows.Close() is critical in the code above. As long as a resultset is open it will keep the underlying database connection open… so if something goes wrong in this method and the resultset isn’t closed, it can rapidly lead to all the connections in our pool being used up.
}
