package cmd

import (
	"log"

	"github.com/arudzitis/addlib/openlibrary"
	"github.com/spf13/cobra"
)

var oldAuthorName string
var newAuthorName string

var bookOlid string
var newTitle string

func init() {
	authorCmd.PersistentFlags().StringVarP(&oldAuthorName, "old", "o", "", "previous name")
	authorCmd.PersistentFlags().StringVarP(&newAuthorName, "new", "n", "", "new name")
	authorCmd.MarkPersistentFlagRequired("old")
	authorCmd.MarkPersistentFlagRequired("new")

	titleCmd.PersistentFlags().StringVarP(&bookOlid, "olid", "o", "", "previous name")
	titleCmd.PersistentFlags().StringVarP(&newTitle, "title", "t", "", "new title")
	titleCmd.MarkPersistentFlagRequired("olid")
	titleCmd.MarkPersistentFlagRequired("title")

	updateCmd.AddCommand(authorCmd)
	updateCmd.AddCommand(titleCmd)

	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update individual records in the database",
}

var authorCmd = &cobra.Command{
	Use:   "author",
	Short: "update individual author records in the database",
	Run: func(cmd *cobra.Command, args []string) {
		runUpdateAuthor()
	},
}

func runUpdateAuthor() {
	rows, err := database.UpdateAuthorName(oldAuthorName, newAuthorName)
	cobra.CheckErr(err)

	log.Printf("Updated %d rows!\n", rows)
}

var titleCmd = &cobra.Command{
	Use:   "title",
	Short: "update individual book title records in the database",
	Run: func(cmd *cobra.Command, args []string) {
		runUpdateTitle()
	},
}

func runUpdateTitle() {
	rows, err := database.UpdateTitle(openlibrary.Book{OLID: bookOlid}, newTitle)
	cobra.CheckErr(err)

	log.Printf("Updated %d rows!\n", rows)
}
