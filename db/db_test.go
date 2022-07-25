package db

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/arudzitis/addlib/openlibrary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertRecord(t *testing.T) {

	authorA := openlibrary.Author{
		OLID: "olid-authora",
		Name: "Author A",
	}

	authorB := openlibrary.Author{
		OLID: "olid-authorb",
		Name: "Author B",
	}

	authorC := openlibrary.Author{
		OLID: "olid-authorc",
		Name: "Author C",
	}

	authorD := openlibrary.Author{
		OLID: "olid-authord",
		Name: "Author D",
	}

	db := openTestDatabase(t)
	defer db.Close()

	testCases := []struct {
		name string
		book openlibrary.Book
	}{
		{
			"book with new author",
			openlibrary.Book{
				OLID:    "olid-booka",
				Title:   "Book A",
				Authors: []openlibrary.Author{authorA},
			},
		},
		{
			"book with existing + new",
			openlibrary.Book{
				OLID:    "olid-bookb",
				Title:   "Book B",
				Authors: []openlibrary.Author{authorA, authorB},
			},
		},
		{
			"book with existing + existing",
			openlibrary.Book{
				OLID:    "olid-bookd",
				Title:   "Book D",
				Authors: []openlibrary.Author{authorA, authorB},
			},
		},
		{
			"book with new + new",
			openlibrary.Book{
				OLID:    "olid-booke",
				Title:   "Book E",
				Authors: []openlibrary.Author{authorC, authorD},
			},
		},
	}

	for _, testCase := range testCases {
		err := db.InsertRecord(testCase.book)
		require.NoError(t, err)
	}

	retrivedBooks, err := db.AllBooks()
	require.NoError(t, err)

	require.Len(t, retrivedBooks, len(testCases))

	for i, testCase := range testCases {
		assert.True(t, reflect.DeepEqual(testCase.book, retrivedBooks[i]), "Book for test %q was not equal: (expected %v but was %v)", testCase.name, testCase.book, retrivedBooks[i])

		err := db.InsertRecord(testCase.book)
		require.NoError(t, err)
	}

}

func TestUpdateAuthorName(t *testing.T) {
	authorA := openlibrary.Author{
		OLID: "olid-authora",
		Name: "Author A",
	}

	authorB := openlibrary.Author{
		OLID: "olid-authorb",
		Name: "Author B",
	}

	book := openlibrary.Book{
		OLID:    "olid-booka",
		Title:   "Book A",
		Authors: []openlibrary.Author{authorA, authorB},
	}

	db := openTestDatabase(t)
	defer db.Close()

	err := db.InsertRecord(book)
	require.NoError(t, err)

	rows, err := db.UpdateAuthorName("Author A", "Author C")
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	books, err := db.AllBooks()
	require.NoError(t, err)

	require.Equal(t, 1, len(books))
	assert.Equal(t, "Author C", books[0].Authors[0].Name)
}

func TestUpdateTitle(t *testing.T) {
	book := openlibrary.Book{
		OLID:    "olid-booka",
		Title:   "Book A",
		Authors: []openlibrary.Author{},
	}

	db := openTestDatabase(t)
	defer db.Close()

	err := db.InsertRecord(book)
	require.NoError(t, err)

	rows, err := db.UpdateTitle(openlibrary.Book{OLID: "olid-booka"}, "Book B")
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	books, err := db.AllBooks()
	require.NoError(t, err)

	require.Equal(t, 1, len(books))
	assert.Equal(t, "Book B", books[0].Title)
}

func TestDeleteBook(t *testing.T) {
	authorA := openlibrary.Author{
		OLID: "olid-authora",
		Name: "Author A",
	}

	bookA := openlibrary.Book{
		OLID:    "olid-booka",
		Title:   "Book A",
		Authors: []openlibrary.Author{authorA},
	}

	bookB := openlibrary.Book{
		OLID:    "olid-bookb",
		Title:   "Book B",
		Authors: []openlibrary.Author{authorA},
	}

	db := openTestDatabase(t)
	defer db.Close()

	err := db.InsertRecord(bookA)
	require.NoError(t, err)

	err = db.InsertRecord(bookB)
	require.NoError(t, err)

	rows, err := db.DeleteBook(bookA)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	books, err := db.AllBooks()
	require.NoError(t, err)

	require.Equal(t, 1, len(books))
	assert.Equal(t, "Book B", books[0].Title)
	assert.Equal(t, "Author A", books[0].Authors[0].Name)
}

func openTestDatabase(t *testing.T) *DB {
	t.Helper()

	tempFile, err := ioutil.TempFile("", "*.sqlite3")
	require.NoError(t, err)

	db, err := OpenDatabase(tempFile.Name(), false)
	require.NoError(t, err)

	err = db.Migrate()
	require.NoError(t, err)

	return db
}
