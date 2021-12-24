/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"strings"

	internal "aaronromeo/mailboxorg/caduceus/internal"

	"github.com/spf13/cobra"
)

const labelsArg string = "labels"
const filtersArg string = "filters"

var validArgs = []string{labelsArg, filtersArg, "all"}

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   fmt.Sprintf("fetch [%s]", strings.Join(validArgs, "|")),
	Short: "Fetch Gmail resources",
	Long: `Usage:
fetch
fetch all
fetch labels
fetch filters`,
	ValidArgs: validArgs,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}

		if contains(validArgs, args[0]) {
			return nil
		}

		return fmt.Errorf("invalid args: %s", args[0])
	},
	Run: fetchLabelsAndFilters,
}

func fetchLabelsAndFilters(cmd *cobra.Command, args []string) {
	switch args[0] {
	case labelsArg:
		fetchLabels()
	case filtersArg:
		fetchFilters()
	default:
		fetchLabels()
		fetchFilters()
	}
}

func fetchLabels() {
	fmt.Println("Fetching labels...")
	labels, err := internal.GetLabels()
	if err != nil {
		panic(err)
	}
	err = internal.SaveLocalLabels(labels)
	if err != nil {
		panic(err)
	}
}

func fetchFilters() {
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

func contains(s []string, str string) bool {
	for _, v := range validArgs {
		if v == str {
			return true
		}
	}

	return false
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// fetchCmd.Flags().StringVarP(&Resource, "resource", "r", "all", "Resources:labels,filters,all")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
