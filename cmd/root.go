package cmd

import (
	"log"

	"github.com/arudzitis/addlib/db"
	"github.com/spf13/cobra"
)

var (
	databaseFile string
	database     *db.DB

	verbose bool

	rootCmd = &cobra.Command{
		Use:   "addlib",
		Short: "addlib is tool for managing an inventory of a small library, using data from openlibrary.org",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initDatabase)

	rootCmd.PersistentFlags().StringVarP(&databaseFile, "database", "d", "", "database file to use; will be created if it does not exist")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.MarkPersistentFlagRequired("database")
}

func initDatabase() {
	var err error
	database, err = db.OpenDatabase(databaseFile, verbose)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
