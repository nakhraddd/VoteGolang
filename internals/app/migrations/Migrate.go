package migrations

import (
	candidate_data2 "VoteGolang/internals/domain"
	"VoteGolang/internals/infrastructure/security"
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
		{"user", &candidate_data2.User{}},
		{"candidate", &candidate_data2.Candidate{}},
		{"petition", &candidate_data2.Petition{}},
		{"petition_vote", &candidate_data2.PetitionVote{}},
		{"petition", &candidate_data2.Vote{}},
		{"role", &candidate_data2.Role{}},
		{"access", &candidate_data2.Access{}},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m.model); err != nil {
			return err
		}
		if err := writeMigrationSummaryFromModel(db, m.name, m.model); err != nil {
			log.Printf("Failed to write migration file for %s: %v", m.name, err)
		}
	}

	seedRBAC(db)

	err := SeedAdminUser(db)
	if err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}

	log.Println("Database tables migrated successfully and migration summaries saved.")
	return nil
}

func SeedAdminUser(db *gorm.DB) error {
	var role candidate_data2.Role
	if err := db.Where("name = ?", "admin").First(&role).Error; err != nil {
		return fmt.Errorf("admin role not found: %w", err)
	}

	password, err := security.HashPassword("admin123")
	if err != nil {
		return err
	}

	admin := candidate_data2.User{
		Username:      "admin",
		Email:         "admin@example.com",
		Password:      password,
		RoleID:        role.ID,
		EmailVerified: true,
	}

	db.FirstOrCreate(&admin, candidate_data2.User{Email: admin.Email})
	return nil
}

func seedRBAC(db *gorm.DB) {
	// These have to be changed
	accesses := []candidate_data2.Access{
		{Name: "create_candidate"},
		{Name: "read_candidate"},
		{Name: "update_candidate"},
		{Name: "delete_candidate"},
		{Name: "create_user"},
		{Name: "read_user"},
		{Name: "update_user"},
		{Name: "delete_user"},
		{Name: "create_petition"},
		{Name: "read_petition"},
		{Name: "update_petition"},
		{Name: "delete_petition"},
		{Name: "vote"},
	}
	for _, a := range accesses {
		db.FirstOrCreate(&a, candidate_data2.Access{Name: a.Name})
	}

	roles := []candidate_data2.Role{
		{Name: "admin"},
		{Name: "member"},
		{Name: "guest"},
	}
	for _, r := range roles {
		db.FirstOrCreate(&r, candidate_data2.Role{Name: r.Name})
	}

	// assign role->access relationships
	var admin, member, guest candidate_data2.Role
	db.First(&admin, "name = ?", "admin")
	db.First(&member, "name = ?", "member")
	db.First(&guest, "name = ?", "guest")

	var create_candidate, read_candidate, update_candidate, delete_candidate, vote,
		create_petition, read_petition, update_petition, delete_petition candidate_data2.Access
	db.First(&create_candidate, "name = ?", "create_candidate")
	db.First(&read_candidate, "name = ?", "read_candidate")
	db.First(&update_candidate, "name = ?", "update_candidate")
	db.First(&delete_candidate, "name = ?", "delete_candidate")

	db.First(&vote, "name = ?", "vote")

	db.First(&create_petition, "name = ?", "create_petition")
	db.First(&read_petition, "name = ?", "read_petition")
	db.First(&update_petition, "name = ?", "update_petition")
	db.First(&delete_petition, "name = ?", "delete_petition")

	// These have to be changed
	db.Model(&admin).Association("Accesses").Replace(&[]candidate_data2.Access{
		create_candidate, read_candidate, update_candidate, delete_candidate, create_petition,
		read_petition, update_petition, delete_petition,
	})
	db.Model(&member).Association("Accesses").Replace(&[]candidate_data2.Access{
		read_candidate, vote, create_petition, read_petition, update_petition, delete_petition,
	})
	db.Model(&guest).Association("Accesses").Replace(&[]candidate_data2.Access{})
}
