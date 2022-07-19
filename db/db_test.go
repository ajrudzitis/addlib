package db

import (
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

func openTestDatabase(t *testing.T) *DB {
	t.Helper()

	db, err := OpenDatabase("file::memory:?cache=shared", false)
	require.NoError(t, err)

	err = db.Migrate()
	require.NoError(t, err)

	return db
}
