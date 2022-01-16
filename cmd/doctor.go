/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	internal "aaronromeo/mailboxorg/caduceus/internal"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) {
	FetchLabels()
	// FetchFilters()

	prompt := promptui.Select{
		Label: "Select Day",
		Items: []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday",
			"Saturday", "Sunday"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", result)

	fmt.Println("Analyzing results...")
	err := emptyLabelMigrations()
	criteriaAndSampleMessages, err := internal.GetMessageCriteriaForUnsubscribe(time.Now().Add(-time.Hour * 72).UTC())
	if err != nil {
		panic(err)
	}
	for _, cAndSM := range criteriaAndSampleMessages {
		criteria := *cAndSM.Criteria
		if criteria.Query != "" {
			fmt.Println("query: ", criteria.Query)
		} else if criteria.From != "" {
			fmt.Println("from:", criteria.From)
		} else if criteria.To != "" {
			fmt.Println("to:", criteria.To)
		}
		subject := ""
		for _, header := range (*cAndSM.SampleMessage).Payload.Headers {
			if header.Name == "Subject" && header.Value != "" {
				subject = header.Value
			}

		}
		fmt.Println("\t\tmessage: ", subject)
	}
}

func emptyLabel(l internal.CadLabel) bool {
	// For some reason MessagesTotal and ThreadsTotal come back empty
	// in spite of there being messages
	return l.MessagesTotal == 0 &&
		l.MessagesUnread == 0 &&
		l.ThreadsTotal == 0 &&
		l.ThreadsUnread == 0
}

func emptyLabelMigrations() error {
	localLabels, err := internal.ReadLocalLabels()
	if err != nil {
		log.Printf("Unable to read local labels: %v", err)
		return err
	}

	emptyLabels := []internal.CadLabel{}
	nestedLabelLookup := map[string][]internal.CadLabel{}

	for _, label := range localLabels {
		labelParts := strings.Split(label.Name, "/")
		for i := range labelParts {
			key := strings.Join(labelParts[:i+1], "/")
			nestedLabelLookup[key] = append(nestedLabelLookup[key], label)
		}
		if emptyLabel(label) {
			emptyLabels = append(emptyLabels, label)
		}
	}

	emptyLabelMigrations := []internal.CadRawMigration{}
	for _, label := range emptyLabels {
		if len(nestedLabelLookup[label.Name]) == 1 {
			operation := internal.DeleteLabelMigration
			note := fmt.Sprintf("%s: Empty Label identified by the doctor", label.Name)
			labelId := label.Id
			labelMigration := internal.CadRawMigration{
				Operation: &operation,
				Details: internal.CadDeleteLabelMigration{
					Id: &labelId,
				},
				Note: &note,
			}

			emptyLabelMigrations = append(emptyLabelMigrations, labelMigration)
		} else {
			fmt.Printf("Skipping parent label %s\n", label.Name)
		}
	}
	return internal.CreateMigrationFile(&emptyLabelMigrations)
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// doctorCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// doctorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
