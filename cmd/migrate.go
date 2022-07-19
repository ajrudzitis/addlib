package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate the database schema",
	Run: func(cmd *cobra.Command, args []string) {
		runMigrate()
	},
}

func runMigrate() {
	err := database.Migrate()
	cobra.CheckErr(err)
}
