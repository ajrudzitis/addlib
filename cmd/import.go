package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/arudzitis/addlib/openlibrary"
	"github.com/spf13/cobra"
)

var inputFileName string
var inputFormatName string
var exceptionFileName string

func init() {
	importCmd.PersistentFlags().StringVarP(&inputFileName, "input", "i", "", "file to import from")
	importCmd.PersistentFlags().StringVarP(&inputFormatName, "format", "f", "", `"isbn" or "olid" for openlibrary id`)
	importCmd.PersistentFlags().StringVarP(&exceptionFileName, "exceptions", "e", "", "file to write lines which were not able to be imported")
	importCmd.MarkPersistentFlagRequired("input")
	importCmd.MarkPersistentFlagRequired("format")

	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import new data into the database",
	Run: func(cmd *cobra.Command, args []string) {
		runImport()
	},
}

func runImport() {
	inputFile, err := os.Open(inputFileName)
	cobra.CheckErr(err)
	defer func() { _ = inputFile.Close() }()

	var exceptionFile *os.File
	if exceptionFileName != "" {
		exceptionFile, err = os.Create(exceptionFileName)
		cobra.CheckErr(err)
		defer func() { _ = exceptionFile.Close() }()
	}

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
			if exceptionFile != nil {
				_, err = exceptionFile.Write([]byte(fmt.Sprintf("%s\n", nextLine)))
				cobra.CheckErr(err)
			}
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
