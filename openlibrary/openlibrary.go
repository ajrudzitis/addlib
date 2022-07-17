package openlibrary

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const isbnSeparator = ","

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

func (b *Book) GetIsbn13() *string {
	if len(b.Isbn13) == 0 {
		return nil
	}
	result := strings.Join(b.Isbn13, isbnSeparator)
	return &result
}

func (b *Book) GetIsbn10() *string {
	if len(b.Isbn10) == 0 {
		return nil
	}
	result := strings.Join(b.Isbn10, isbnSeparator)
	return &result
}

func (b *Book) SetIsbn13(input string) {
	b.Isbn13 = strings.Split(input, isbnSeparator)
}

func (b *Book) SetIsbn10(input string) {
	b.Isbn10 = strings.Split(input, isbnSeparator)
}

type Author struct {
	OLID string `json:"key"`
	Name string `json:"name"`
}

func LookupByISBN(isbn string) (*Book, error) {
	response, err := http.Get(fmt.Sprintf("https://openlibrary.org/isbn/%s.json", isbn))
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error making request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openlibrary: non 200 stats while looking up isbn: %s", isbn)
	}

	result := &Book{}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error reading response: %w", err)
	}

	err = json.Unmarshal(responseBody, result)
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error unmarshaling isbn response: %w", err)
	}

	// The quality of title string and authors in the parent work object seems to be better, so use
	// that if it is present and unambiguous
	if len(result.Works) == 1 {
		work, err := lookupWorkByKey(result.Works[0].Key)
		if err != nil {
			return nil, err
		}
		result.Title = work.Title
		if len(work.Authors) > 0 {
			authors := []Author{}
			for _, a := range work.Authors {
				authors = append(authors, a.Author)
			}
			result.Authors = authors
		}
	}

	for i := range result.Authors {
		resolvedAuthor, err := lookupAuthorByKey(result.Authors[i].OLID)
		if err != nil {
			return nil, err
		}
		result.Authors[i] = *resolvedAuthor
	}

	return result, nil
}

func lookupAuthorByKey(key string) (*Author, error) {
	response, err := http.Get(fmt.Sprintf("https://openlibrary.org/%s.json", key))
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error making request: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openlibrary: non 200 stats while looking up author: %s", key)
	}

	result := &Author{}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error reading response: %w", err)
	}

	err = json.Unmarshal(responseBody, result)
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error unmarshaling author response: %w", err)
	}

	return result, nil
}

type work struct {
	Title   string `json:"title"`
	Authors []struct {
		Author `json:"author"`
	} `json:"authors"`
}

func lookupWorkByKey(key string) (*work, error) {
	response, err := http.Get(fmt.Sprintf("https://openlibrary.org/%s.json", key))
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error making request: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openlibrary: non 200 stats while looking up work: %s", key)
	}

	result := &work{}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error reading response: %w", err)
	}

	err = json.Unmarshal(responseBody, result)
	if err != nil {
		return nil, fmt.Errorf("openlibrary: error unmarshaling works response: %w", err)
	}

	return result, nil
}
