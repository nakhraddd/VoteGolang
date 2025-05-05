package migrations

import (
	"VoteGolang/internals/data/candidate_data"
	"VoteGolang/internals/data/news_data"
	"VoteGolang/internals/data/petition_data"
	"VoteGolang/internals/data/user_data"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

func writeMigrationSummaryFromModel(db *gorm.DB, modelName string, model interface{}) error {
	dir := "../../internals/app/migrations/result"
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("package result\n\n")

	sb.WriteString(fmt.Sprintf("// Migration summary for %s\n", modelName))
	sb.WriteString(fmt.Sprintf("// Table: %s\n", stmt.Table))
	sb.WriteString("// -----------------------------------\n")
	sb.WriteString(fmt.Sprintf("// CREATE TABLE %s (\n", stmt.Table))

	for _, field := range stmt.Schema.Fields {
		sb.WriteString("//   " + field.DBName + " ")

		// Try to get explicit type from tag
		fieldType := extractTypeFromTag(field.Tag.Get("gorm"))
		if fieldType == "" {
			// fallback to field.DataType (basic string like "string", "int", etc.)
			fieldType = string(field.DataType)
		}
		sb.WriteString(fieldType)

		// Constraints
		if field.NotNull {
			sb.WriteString(" NOT NULL")
		}
		if field.PrimaryKey {
			sb.WriteString(" PRIMARY KEY")
		}
		if field.AutoIncrement {
			sb.WriteString(" AUTO_INCREMENT")
		}
		if field.Unique {
			sb.WriteString(" UNIQUE")
		}
		if field.DefaultValueInterface != nil {
			sb.WriteString(fmt.Sprintf(" DEFAULT '%v'", field.DefaultValueInterface))
		}

		sb.WriteString(",\n")
	}

	sb.WriteString("// );\n")
	sb.WriteString("// -----------------------------------\n")

	filePath := filepath.Join(dir, fmt.Sprintf("%s.go", modelName))
	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}

func extractTypeFromTag(tag string) string {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "type:") {
			return strings.TrimPrefix(part, "type:")
		}
	}
	return ""
}

func MigrateAllTables(db *gorm.DB) error {
	models := []struct {
		name  string
		model interface{}
	}{
		{"user", &user_data.User{}},
		{"candidate", &candidate_data.Candidate{}},
		{"general_news", &news_data.GeneralNews{}},
		{"petition_data", &petition_data.Petition{}},
		{"petition_vote", &petition_data.PetitionVote{}},
		{"petition_data", &candidate_data.Vote{}},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m.model); err != nil {
			return err
		}
		if err := writeMigrationSummaryFromModel(db, m.name, m.model); err != nil {
			log.Printf("Failed to write migration file for %s: %v", m.name, err)
		}
	}

	log.Println("Database tables migrated successfully and migration summaries saved.")
	return nil
}
