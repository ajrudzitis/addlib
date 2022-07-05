package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"

	openlibrary "github.com/arudzitis/addlib/openlibary"
	"github.com/spf13/cobra"
)

var inputFileName string
var inputFormatName string

func init() {
	importCmd.PersistentFlags().StringVarP(&inputFileName, "input", "i", "", "file to import from")
	importCmd.PersistentFlags().StringVarP(&inputFormatName, "format", "f", "", `"isbn" or "olid" for openlibrary id`)
	importCmd.MarkPersistentFlagRequired("input")
	importCmd.MarkPersistentFlagRequired("format")

	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import new data into the database",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func run() {
	inputFile, err := os.Open(inputFileName)
	cobra.CheckErr(err)
	defer func() { _ = inputFile.Close() }()

	var handler func(string) error

	switch inputFormatName {
	case "isbn":
		handler = handleIsbn
	case "olid":
		handler = handleOlid
	default:
		log.Fatalf("unsupported input format: %q", inputFormatName)
	}

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		nextLine := scanner.Text()
		err = handler(nextLine)
		if err != nil {
			log.Printf("error handling input %q; %v, skipping...", nextLine, err)
		}
	}
}

func handleIsbn(isbn string) error {
	cleanedIsbn, err := sanitizeISBN(isbn)
	if err != nil {
		return err
	}

	book, err := openlibrary.LookupByISBN(cleanedIsbn)
	if err != nil {
		return err
	}

	err = database.InsertRecord(*book)
	if err != nil {
		return err
	}

	return nil
}

func handleOlid(olid string) error {
	return nil
}

var (
	isbn13Pattern = regexp.MustCompile(`(^97[\d]{11}).*`)
	isbn10Pattern = regexp.MustCompile(`(^[\d]{10}).*`)
)

func sanitizeISBN(isbn string) (string, error) {
	if match := isbn13Pattern.FindStringSubmatch(isbn); match != nil {
		return match[1], nil
	}

	if match := isbn10Pattern.FindStringSubmatch(isbn); match != nil {
		return match[1], nil
	}

	return "", fmt.Errorf("%s does not appear to be an isbn13 or isbn10", isbn)
}
