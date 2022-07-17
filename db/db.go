package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/arudzitis/addlib/openlibrary"
	_ "github.com/mattn/go-sqlite3"
)

const createBooksQuery = "CREATE TABLE IF NOT EXISTS books(id INTEGER PRIMARY KEY AUTOINCREMENT, olid TEXT, isbn13 TEXT, isbn10 TEXT, title TEXT NOT NULL, override_title TEXT);"
const createAuthorsQuery = "CREATE TABLE IF NOT EXISTS authors(id INTEGER PRIMARY KEY AUTOINCREMENT, olid TEXT, name TEXT NOT NULL);"
const createBookAuthorsQuery = "CREATE TABLE IF NOT EXISTS book_authors(book INTEGER, author INTEGER, FOREIGN KEY(book) REFERENCES books(id), FOREIGN KEY(author) REFERENCES authors(id));"

const insertBookQuery = "INSERT INTO books(olid, isbn13, isbn10, title) values (?,?,?,?);"
const insertAuthorQuery = "INSERT INTO authors(olid, name) values (?,?);"
const insertBookAuthorQuery = "INSERT INTO book_authors(book, author) values (?,?);"

const readBookByOLIDQuery = "SELECT olid, isbn13, isbn10, title FROM books where olid = ?;"
const readAuthorByOLIDQuery = "SELECT id, olid, name FROM authors where olid = ?;"
const readAuthorsByBookOLIDQuery = "SELECT authors.olid, authors.name FROM authors INNER JOIN book_authors ON authors.id = book_authors.author WHERE book_authors.book = ?;"

const readAllBooksQuery = "SELECT id, olid, isbn13, isbn10, title, override_title FROM books;"

const updateBookOverrideTitle = "UPDATE books SET override_title = ? WHERE olid = ?;"

type DB struct {
	db *sql.DB
}

func (d DB) Close() {
	d.db.Close()
}

func OpenDatabase(databasePath string) (*DB, error) {
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, fmt.Errorf("db: error opening database: %w", err)
	}

	for _, query := range []string{createBooksQuery, createAuthorsQuery, createBookAuthorsQuery} {
		_, err = db.Exec(query)
		if err != nil {
			return nil, fmt.Errorf("db: error initializing database with statement: %s: %w", query, err)
		}
	}
	return &DB{db: db}, nil
}

func (d DB) InsertRecord(book openlibrary.Book) error {
	// check if book exists
	existingBook, err := d.readBook(book.OLID)
	if err != nil {
		return err
	}
	if existingBook != nil {
		log.Printf("Book %s already saved.\n", book.Title)
		return nil
	}

	// insert the book
	insertBookStmt, err := d.db.Prepare(insertBookQuery)
	if err != nil {
		return err
	}
	defer insertBookStmt.Close()

	res, err := insertBookStmt.Exec(book.OLID, book.GetIsbn13(), book.GetIsbn10(), book.Title)
	if err != nil {
		return fmt.Errorf("db: error inserting book %s: %w", book.OLID, err)
	}
	bookId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("db: error inserting book %s: %w", book.OLID, err)
	}

	// map the author OLIDs to the database ids
	authorOlidToId := map[string]int64{}

	// check if authors exist, insert any missing
	insertAuthorStmt, err := d.db.Prepare(insertAuthorQuery)
	if err != nil {
		return err
	}
	defer insertAuthorStmt.Close()
	for _, author := range book.Authors {
		existingAuthor, authorId, err := d.readAuthor(author.OLID)
		if err != nil {
			return err
		}
		if existingAuthor != nil {
			authorOlidToId[author.OLID] = *authorId
			continue
		}
		res, err = insertAuthorStmt.Exec(author.OLID, author.Name)
		if err != nil {
			return fmt.Errorf("db: error inserting author %s: %w", author.OLID, err)
		}
		newAuthorId, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("db: error inserting author %s: %w", author.OLID, err)
		}
		authorOlidToId[author.OLID] = newAuthorId
	}

	// add the book/author relationships
	insertBookAuthorStmt, err := d.db.Prepare(insertBookAuthorQuery)
	if err != nil {
		return err
	}
	defer insertBookAuthorStmt.Close()
	for _, author := range book.Authors {
		_, err = insertBookAuthorStmt.Exec(bookId, authorOlidToId[author.OLID])
		if err != nil {
			return fmt.Errorf("db: error inserting book author %s-%s: %w", book.OLID, author.OLID, err)
		}
	}
	return nil
}

func (d DB) UpdateTitle(book openlibrary.Book, title string) error {
	stmt, err := d.db.Prepare(updateBookOverrideTitle)
	if err != nil {
		return fmt.Errorf("db: error preparing statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(book.OLID, title)
	if err != nil {
		return fmt.Errorf("db: error overriding title: %w", err)
	}
	return nil
}

func (d DB) AllBooks() ([]openlibrary.Book, error) {
	stmt, err := d.db.Prepare(readAllBooksQuery)
	if err != nil {
		return nil, fmt.Errorf("db: error preparing statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("db: error reading all books: %w", err)
	}

	books := []openlibrary.Book{}
	for {
		if rows.Next() {

			var bookId int64
			book := openlibrary.Book{}
			overrideTitle := sql.NullString{}
			isbn13 := sql.NullString{}
			isbn10 := sql.NullString{}
			err = rows.Scan(&bookId, &book.OLID, &isbn13, &isbn10, &book.Title, &overrideTitle)
			if err != nil {
				return nil, fmt.Errorf("db: error reading book: %w", err)
			}

			if isbn13.Valid {
				book.SetIsbn13(isbn13.String)
			}

			if isbn10.Valid {
				book.SetIsbn10(isbn10.String)
			}

			if overrideTitle.Valid {
				book.Title = overrideTitle.String
			}

			book.Authors, err = d.readBookAuthors(bookId)
			if err != nil {
				return nil, fmt.Errorf("db: error reading authors associated with book: %w", err)
			}

			books = append(books, book)
		} else {
			if rows.Err() != nil {
				return nil, fmt.Errorf("db: error reading all books: %w", err)
			}
			break
		}
	}

	return books, nil
}

func (d DB) readBook(olid string) (*openlibrary.Book, error) {
	stmt, err := d.db.Prepare(readBookByOLIDQuery)
	if err != nil {
		return nil, fmt.Errorf("db: error preparing statement: %w", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(olid)
	if row.Err() != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: error querying books: %w", err)
	}

	book := &openlibrary.Book{}
	isbn13 := sql.NullString{}
	isbn10 := sql.NullString{}
	err = row.Scan(&book.OLID, &isbn13, &isbn10, &book.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("db: error scanning book result: %w", err)
	}

	if isbn13.Valid {
		book.SetIsbn13(isbn13.String)
	}

	if isbn10.Valid {
		book.SetIsbn10(isbn10.String)
	}

	return book, nil
}

func (d DB) readAuthor(olid string) (*openlibrary.Author, *int64, error) {
	stmt, err := d.db.Prepare(readAuthorByOLIDQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("db: error preparing statement: %w", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow(olid)
	if row.Err() != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("db: error querying authors: %w", err)
	}
	author := &openlibrary.Author{}
	var authorId int64
	err = row.Scan(&authorId, &author.OLID, &author.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("db: error scanning author result: %w", err)
	}

	return author, &authorId, nil
}

func (d DB) readBookAuthors(bookID int64) ([]openlibrary.Author, error) {
	stmt, err := d.db.Prepare(readAuthorsByBookOLIDQuery)
	if err != nil {
		return nil, fmt.Errorf("db: error preparing statement: %w", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(bookID)
	if err != nil {
		return nil, fmt.Errorf("db: error querying authors: %w", err)
	}
	authors := []openlibrary.Author{}

	for {
		if rows.Next() {
			author := openlibrary.Author{}
			err = rows.Scan(&author.OLID, &author.Name)
			if err != nil {
				return nil, fmt.Errorf("db: error scanning author result: %w", err)
			}
			authors = append(authors, author)
		} else {
			if rows.Err() != nil {
				return nil, fmt.Errorf("db: error reading book authors: %w", err)
			}
			break
		}
	}
	return authors, nil
}
