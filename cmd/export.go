package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var outputFileName string

func init() {
	exportCmd.PersistentFlags().StringVarP(&outputFileName, "output", "o", "", "file to output to")
	exportCmd.MarkPersistentFlagRequired("output")

	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "output data from the database",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func run() {
	outputFile, err := os.Create(outputFileName)
	cobra.CheckErr(err)
	defer func() { _ = outputFile.Close() }()

	books, err := database.AllBooks()
	cobra.CheckErr(err)

	_, err = outputFile.Write([]byte("title,author,url\n"))
	cobra.CheckErr(err)

	for _, book := range books {
		authorNames := make([]string, len(book.Authors))
		for i, author := range book.Authors {
			authorNames[i] = author.Name
		}

		_, err = outputFile.Write([]byte(fmt.Sprintf(`"%s","%s","https://openlibrary.org%s"`, book.Title, strings.Join(authorNames, ", "), book.OLID)))
		cobra.CheckErr(err)
		_, err = outputFile.Write([]byte("\n"))
		cobra.CheckErr(err)
	}
}
