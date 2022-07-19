package openlibrary

type Author struct {
	OLID string `json:"key"`
	Name string `json:"name"`
}

type Book struct {
	OLID    string   `json:"key"`
	Title   string   `json:"title"`
	Isbn10  []string `json:"isbn_10"`
	Isbn13  []string `json:"isbn_13"`
	Authors []Author `json:"authors"`
	Works   []struct {
		Key string `json:"key"`
	} `json:"works"`
}
