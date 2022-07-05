package cmd

import (
	"log"

	"github.com/arudzitis/addlib/db"
	"github.com/spf13/cobra"
)

var (
	databaseFile string
	database     *db.DB

	rootCmd = &cobra.Command{
		Use:   "addlib",
		Short: "addlib is tool for managing an inventory of a small library, using data from openlibrary.org",
	}
)

func Execute() error {
	// ensure we clean up the database, if it exists
	defer func() {
		if database != nil {
			database.Close()
		}
	}()
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initDatabase)

	rootCmd.PersistentFlags().StringVarP(&databaseFile, "database", "d", "", "database file to use; will be created if it does not exist")
	rootCmd.MarkPersistentFlagRequired("database")
}

func initDatabase() {
	var err error
	database, err = db.OpenDatabase(databaseFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
