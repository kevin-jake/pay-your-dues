package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"
)

const seedMarkerEmail = "alice@dev.local"

// SeedDevelopmentData loads test-data.sql when running in development mode.
// Seeding is skipped if the marker user already exists.
func SeedDevelopmentData(db *gorm.DB, seedFile string) error {
	var count int64
	if err := db.Table("users").Where("email = ?", seedMarkerEmail).Count(&count).Error; err != nil {
		return fmt.Errorf("check seed status: %w", err)
	}
	if count > 0 {
		log.Println("Development seed data already present, skipping")
		return nil
	}

	content, resolvedPath, err := readSeedFile(seedFile)
	if err != nil {
		return err
	}

	log.Printf("Seeding development data from %s...", resolvedPath)

	for _, statement := range splitSQLStatements(string(content)) {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("execute seed SQL: %w", err)
		}
	}

	log.Println("Development seed data loaded successfully")
	return nil
}

func readSeedFile(seedFile string) ([]byte, string, error) {
	candidates := []string{}
	if seedFile != "" {
		candidates = append(candidates, seedFile)
	}
	candidates = append(candidates, "test-data.sql", "./test-data.sql")

	for _, path := range candidates {
		content, err := os.ReadFile(path)
		if err == nil {
			return content, path, nil
		}
	}

	return nil, "", fmt.Errorf("seed file not found (set SEED_DATA_FILE or place test-data.sql in the project root)")
}

func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inDollarQuote := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		if strings.Contains(line, "$$") {
			count := strings.Count(line, "$$")
			if count%2 == 1 {
				inDollarQuote = !inDollarQuote
			}
		}

		current.WriteString(line)
		current.WriteByte('\n')

		if !inDollarQuote && strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" && stmt != ";" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	if tail := strings.TrimSpace(current.String()); tail != "" {
		statements = append(statements, tail)
	}

	return statements
}
