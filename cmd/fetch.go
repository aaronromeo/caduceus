/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	internal "aaronromeo/mailboxorg/caduceus/internal"

	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch Gmail resources",
	Long: `Usage:
fetch labels
fetch filters`,
	Run: fetchLabelsAndFilters,
}

func fetchLabelsAndFilters(cmd *cobra.Command, args []string) {
	fmt.Println("Fetching labels...")
	labels, err := internal.GetUserLabels()
	if err != nil {
		panic(err)
	}
	err = internal.SaveLocalLabels(labels)
	if err != nil {
		panic(err)
	}

	fmt.Println("Fetching filters...")
	filters, err := internal.GetFilters()
	if err != nil {
		panic(err)
	}
	err = internal.SaveLocalFilters(filters)
	if err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
