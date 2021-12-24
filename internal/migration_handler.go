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

type CadMessagesMigrationAction struct {
	QueryLabelIds  []string `json:"queryLabelIds"`
	RemoveLabelIds []string `json:"removeLabelIds"`
	AddLabelIds    []string `json:"addLabelIds"`
}

type CadMessagesMigration struct {
	Resource string                     `json:"resource"`
	Action   CadMessagesMigrationAction `json:"action"`
}

type CadFiltersMigrationAction struct {
	Id                string   `json:"id"`
	UpdateAddLabelIds []string `json:"updateAddLabelIds"`
}

type CadFiltersMigration struct {
	Resource string                    `json:"resource"`
	Action   CadFiltersMigrationAction `json:"action"`
}

type CadRawMigration struct {
	Resource  string          `json:"resource"`
	Action    interface{}     `json:"-"`
	RawAction json.RawMessage `json:"action"`
}

const migrationsPath string = "migrations"

func RunMigrations() error {
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		log.Fatalf("Unable to fetch migration files: %v", err)
		return err
	}

	var migrations []CadRawMigration
	for _, migrationFile := range migrationFiles {
		b, err := ioutil.ReadFile(migrationFile)
		if err != nil {
			log.Fatalf("Unable to read the migration file: %v", err)
			return err
		}
		if err := json.Unmarshal(b, &migrations); err != nil {
			return err
		}
	}

	for _, migration := range migrations {
		switch migration.Resource {
		case "messages":
			messageMigration := CadMessagesMigration{}
			b, _ := migration.MarshalJSON()
			json.Unmarshal(b, &messageMigration)
			fmt.Println(
				"messageMigration",
				messageMigration.Resource,
				messageMigration.Action.QueryLabelIds,
				messageMigration.Action.AddLabelIds,
				messageMigration.Action.RemoveLabelIds,
			)
			err := migrateMessage(messageMigration)
			if err != nil {
				return err
			}
		case "filters":
			filterMigration := CadFiltersMigration{}
			b, _ := migration.MarshalJSON()
			json.Unmarshal(b, &filterMigration)
			fmt.Println(
				"filterMigration",
				filterMigration.Resource,
				filterMigration.Action.Id,
				filterMigration.Action.UpdateAddLabelIds,
			)
		default:
			return errors.New("Unknown resource " + migration.Resource)
		}
	}

	return nil
}

func (m *CadRawMigration) UnmarshalJSON(b []byte) error {
	type cadRawMigration CadRawMigration

	err := json.Unmarshal(b, (*cadRawMigration)(m))
	if err != nil {
		log.Fatalf("Unable to marshal data: %v", err)
		return err
	}

	return nil
}

func (m *CadRawMigration) MarshalJSON() ([]byte, error) {
	type cadRawMigration CadRawMigration

	if m.Action != nil {
		b, err := json.Marshal(m.Action)

		if err != nil {
			return nil, err
		}
		m.RawAction = b
	}
	return json.Marshal((*cadRawMigration)(m))
}

func getMigrationFiles() ([]string, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		log.Fatalf("Unable to read the migrations directory: %v", err)
		return nil, err
	}

	migrationFiles := []string{}
	for _, file := range files {
		r, _ := regexp.Compile("[0-9]+.json$")
		if r.MatchString(file.Name()) {
			migrationFiles = append(migrationFiles, strings.Join([]string{migrationsPath, file.Name()}, "/"))
		}
	}
	sort.SliceStable(migrationFiles, func(i, j int) bool {
		return migrationFiles[i] < migrationFiles[j]
	})

	return migrationFiles, nil
}

func migrateMessage(migration CadMessagesMigration) error {
	fmt.Println("Fetching messages...")

	labels := []*CadLabel{}
	for _, labelId := range migration.Action.QueryLabelIds {
		labels = append(labels, &CadLabel{Id: labelId})
	}
	messageIDs, err := GetMessagesIDsByLableIDs(labels)
	if err != nil {
		return err
	}
	fmt.Println(len(messageIDs), messageIDs)

	return nil
}
