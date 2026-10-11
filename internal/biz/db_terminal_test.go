package biz

import (
	"context"
	"strings"
	"testing"
)

func TestDatabaseTerminalPostgreSQLMetaCommands(t *testing.T) {
	connection := &DatabaseTerminalConnection{databaseType: "postgresql"}
	tests := []struct {
		command string
		query   string
		message string
	}{
		{command: "\\l", query: "FROM pg_database"},
		{command: "\\dt", query: "FROM pg_catalog.pg_tables"},
		{command: "\\d users", query: "table_name = 'users'"},
		{command: "\\dn", query: "FROM pg_catalog.pg_namespace"},
		{command: "\\?", message: "Supported commands"},
		{command: "\\c", message: "Usage"},
		{command: "\\schema", message: "Usage"},
	}
	for _, test := range tests {
		query, message, handled := connection.ResolveMetaCommand(context.Background(), test.command)
		if !handled {
			t.Fatalf("expected %q to be handled", test.command)
		}
		if test.query != "" && !strings.Contains(query, test.query) {
			t.Fatalf("expected query for %q to contain %q, got %q", test.command, test.query, query)
		}
		if test.message != "" && !strings.Contains(message, test.message) {
			t.Fatalf("expected message for %q to contain %q, got %q", test.command, test.message, message)
		}
	}
}

func TestDatabaseTerminalLeavesSQLUnchanged(t *testing.T) {
	connection := &DatabaseTerminalConnection{databaseType: "postgresql"}
	if _, _, handled := connection.ResolveMetaCommand(context.Background(), "SELECT 1"); handled {
		t.Fatal("expected ordinary SQL not to be handled as a meta command")
	}
}

func TestQuotePostgreSQLIdentifier(t *testing.T) {
	if got := quotePostgreSQLIdentifier(`tenant"one`); got != `"tenant""one"` {
		t.Fatalf("unexpected quoted identifier: %s", got)
	}
}
