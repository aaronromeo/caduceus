/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	internal "aaronromeo/mailboxorg/caduceus/internal"

	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run a pending migration file located in the 'migrations' folder",
	Long:  `Run a pending migration file located in the 'migrations' folder`,
	Run:   runMigrations,
}

func runMigrations(cmd *cobra.Command, args []string) {
	defer func() {
		FetchLabels()
		FetchFilters()
	}()

	err := internal.RunMigrations()
	if err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
