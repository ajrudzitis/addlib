package db

import (
	_ "gorm.io/gorm"
)

type Book struct {
	ID      int64    `gorm:"primaryKey;column:id"`
	OLID    string   `gorm:"index;unique;column:olid;not null"`
	ISBN13  *string  `gorm:"column:isbn13"`
	ISBN10  *string  `gorm:"column:isbn10"`
	Title   string   `gorm:"column:title;not null"`
	Authors []Author `gorm:"many2many:book_authors;"`
}

type Author struct {
	ID   int64  `gorm:"primaryKey;column:id"`
	OLID string `gorm:"unique;column:olid;not null"`
	Name string `gorm:"column:name;not null"`
}

type BookAuthor struct {
	BookID   int64 `gorm:"primaryKey;column:book_id"`
	AuthorID int64 `gorm:"primaryKey;column:author_id"`
}
