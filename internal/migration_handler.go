package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

type CadUpdateMessagesMigration struct {
	QueryLabelIds  *[]string `json:"queryLabelIds"`
	RemoveLabelIds *[]string `json:"removeLabelIds"`
	AddLabelIds    *[]string `json:"addLabelIds"`
}

type CadCreateFilterMigration struct {
	Criteria *CadCriteria `json:"criteria,omitempty"`
	Action   *CadAction   `json:"action,omitempty"`
}

type CadDeleteFilterMigration struct {
	Id *string `json:"id"`
}

type CadRawMigration struct {
	Operation  *string         `json:"operation"`
	Details    interface{}     `json:"-"`
	RawDetails json.RawMessage `json:"details"`
}

const migrationsPath string = "migrations"

func RunMigrations() error {
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		log.Printf("Unable to fetch migration files: %v", err)
		return err
	}

	for _, migrationFile := range migrationFiles {
		var migrations []CadRawMigration

		b, err := ioutil.ReadFile(migrationFile)
		if err != nil {
			log.Printf("Unable to read the migration file: %v", err)
			return err
		}
		if err := json.Unmarshal(b, &migrations); err != nil {
			return err
		}

		for _, migration := range migrations {
			switch *migration.Operation {
			case "update-messages":
				messageMigration := CadUpdateMessagesMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &messageMigration)
				err := updateMessages(messageMigration)
				if err != nil {
					return err
				}
			case "create-filter":
				filterMigration := CadCreateFilterMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &filterMigration)
				err := createFilter(filterMigration)
				if err != nil {
					return err
				}
			case "delete-filter":
				filterMigration := CadDeleteFilterMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &filterMigration)
				err := deleteFilter(filterMigration)
				if err != nil {
					return err
				}
			default:
				return errors.New("Unknown operation " + *migration.Operation)
			}
		}

		err = os.Rename(
			migrationFile,
			strings.ReplaceAll(
				strings.ToLower(migrationFile), ".json", "-complete.json",
			),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *CadRawMigration) UnmarshalJSON(b []byte) error {
	type cadRawMigration CadRawMigration

	err := json.Unmarshal(b, (*cadRawMigration)(m))
	if err != nil {
		log.Printf("Unable to marshal data: %v", err)
		return err
	}

	return nil
}

func (m *CadRawMigration) MarshalJSON() ([]byte, error) {
	type cadRawMigration CadRawMigration

	if m.Details != nil {
		b, err := json.Marshal(m.Details)

		if err != nil {
			return nil, err
		}
		m.RawDetails = b
	}
	return json.Marshal((*cadRawMigration)(m))
}

func getMigrationFiles() ([]string, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		log.Printf("Unable to read the migrations directory: %v", err)
		return nil, err
	}

	migrationFiles := []string{}
	for _, file := range files {
		r, _ := regexp.Compile("[0-9]+.json$")
		if r.MatchString(file.Name()) {
			migrationFiles = append(migrationFiles, strings.Join([]string{migrationsPath, file.Name()}, "/"))
		}
	}
	if len(migrationFiles) == 0 {
		log.Printf("No files to migrate")
		return nil, errors.New("No migrations files")
	}
	sort.SliceStable(migrationFiles, func(i, j int) bool {
		return migrationFiles[i] < migrationFiles[j]
	})

	return migrationFiles, nil
}

func updateMessages(migration CadUpdateMessagesMigration) error {
	fmt.Println("Migrating messages...", migration.QueryLabelIds)

	labels := []*CadLabel{}
	for _, labelId := range *migration.QueryLabelIds {
		labels = append(labels, &CadLabel{Id: labelId})
	}
	messageIds, err := GetMessagesIDsByLabelIDs(labels)
	if err != nil {
		log.Printf("Unable to retrieve message Ids: %v", err)
		return err
	}

	err = BulkUpdateMessageLabels(
		messageIds,
		*migration.AddLabelIds,
		*migration.RemoveLabelIds,
	)
	if err != nil {
		log.Printf("Unable to modify messages: %v", err)
		return err
	}

	return nil
}

func createFilter(migration CadCreateFilterMigration) error {
	fmt.Println("Creating filter...", migration.Criteria.From, migration.Criteria.To, migration.Criteria.Subject)

	newCadFilter := &CadFilter{Action: migration.Action, Criteria: migration.Criteria}

	_, err := CreateFilter(newCadFilter)
	if err != nil {
		log.Printf("Unable to create new filter")
		return err
	}

	return nil
}

func deleteFilter(migration CadDeleteFilterMigration) error {
	fmt.Println("Deleting filter...", *migration.Id)

	oldCadFilter := &CadFilter{Id: *migration.Id}
	oldCadFilter, err := GetFilter(oldCadFilter)
	if err != nil {
		log.Printf("Unable to find referenced filter %v", *migration.Id)
		return err
	}

	err = DeleteFilter(oldCadFilter)
	if err != nil {
		log.Printf("Unable to delete filter %v", *migration.Id)
		return err
	}

	return nil
}
