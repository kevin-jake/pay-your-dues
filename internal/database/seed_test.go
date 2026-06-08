package database

import (
	"strings"
	"testing"
)

func TestSplitSQLStatements(t *testing.T) {
	content := `
-- comment only
DO $$
BEGIN
    INSERT INTO users VALUES (1);
END $$;

INSERT INTO contacts VALUES (1);
`

	statements := splitSQLStatements(content)
	if len(statements) != 2 {
		t.Fatalf("expected 2 statements, got %d: %v", len(statements), statements)
	}

	if !strings.Contains(statements[0], "DO $$") {
		t.Fatalf("first statement should contain DO block: %s", statements[0])
	}

	if !strings.HasPrefix(strings.TrimSpace(statements[1]), "INSERT INTO contacts") {
		t.Fatalf("second statement should be contacts insert: %s", statements[1])
	}
}
