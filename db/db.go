package db

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/arudzitis/addlib/openlibrary"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const updateBookOverrideTitle = "UPDATE books SET override_title = ? WHERE olid = ?;"

type DB struct {
	db *gorm.DB
}

func OpenDatabase(databasePath string, verbose bool) (*DB, error) {
	config := &gorm.Config{}

	if verbose {
		config.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			},
		)
	}

	db, err := gorm.Open(sqlite.Open(databasePath), config)
	if err != nil {
		return nil, fmt.Errorf("db: error opening database: %w", err)
	}

	return &DB{db: db}, nil
}

func (d DB) Migrate() error {
	err := d.db.AutoMigrate(&Book{}, &Author{}, &BookAuthor{})
	if err != nil {
		return fmt.Errorf("db: error updating database file: %w", err)
	}
	return nil
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

	// prepare the object for insertion
	ormAuthors := []Author{}
	for _, author := range book.Authors {
		ormAuthor, err := d.readAuthor(author.OLID)
		if err != nil {
			return err
		}

		if ormAuthor == nil {
			ormAuthor = &Author{
				OLID: author.OLID,
				Name: author.Name,
			}
		}

		ormAuthors = append(ormAuthors, *ormAuthor)
	}

	ormBook := &Book{
		ISBN10:  book.GetIsbn10(),
		ISBN13:  book.GetIsbn13(),
		OLID:    book.OLID,
		Authors: ormAuthors,
		Title:   book.Title,
	}

	tx := d.db.Create(&ormBook)
	if tx.Error != nil {
		return fmt.Errorf("db: error creating book: %w", tx.Error)
	}
	tx = d.db.Save(&ormBook)
	if tx.Error != nil {
		return fmt.Errorf("db: error saving book: %w", tx.Error)
	}

	return nil
}

func (d DB) UpdateTitle(book openlibrary.Book, title string) error {
	tx := d.db.Model(&Book{}).Where("olid = ?", book.OLID).Update("title", title)
	if tx.Error != nil {
		return fmt.Errorf("db: error updating book title: %w", tx.Error)
	}
	book.Title = title

	return nil
}

func (d DB) AllBooks() ([]openlibrary.Book, error) {
	ormBooks := []Book{}
	tx := d.db.Model(&Book{}).Preload("Authors").Find(&ormBooks)
	if tx.Error != nil {
		return nil, fmt.Errorf("db: error reading all books: %w", tx.Error)
	}

	books := []openlibrary.Book{}
	for _, ormBook := range ormBooks {
		ormAuthors, err := d.readAuthors(&ormBook)
		if err != nil {
			return nil, err
		}
		authors := []openlibrary.Author{}
		for _, ormAuthor := range ormAuthors {
			author := openlibrary.Author{
				Name: ormAuthor.Name,
				OLID: ormAuthor.OLID,
			}
			authors = append(authors, author)
		}

		book := openlibrary.Book{
			Title:   ormBook.Title,
			Authors: authors,
			OLID:    ormBook.OLID,
		}

		if ormBook.ISBN10 != nil {
			book.SetIsbn10(*ormBook.ISBN10)
		}

		if ormBook.ISBN13 != nil {
			book.SetIsbn13(*ormBook.ISBN13)
		}

		books = append(books, book)
	}

	return books, nil
}

func (d DB) readBook(olid string) (*Book, error) {
	book := Book{}
	tx := d.db.Find(&book, "olid = ?", olid)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	if tx.RowsAffected != 1 {
		return nil, nil
	}

	return &book, nil
}

func (d DB) readAuthor(olid string) (*Author, error) {
	author := Author{}
	tx := d.db.Find(&author, "olid = ?", olid)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	if tx.RowsAffected != 1 {
		return nil, nil
	}

	return &author, nil
}

func (d DB) readAuthors(book *Book) ([]Author, error) {
	authors := []Author{}
	err := d.db.Model(book).Association("Authors").Find(&authors)
	if err != nil {
		return nil, fmt.Errorf("db: error reading authors for book: %w", err)
	}
	log.Printf("Found %d authors\n", len(authors))
	return authors, nil
}
