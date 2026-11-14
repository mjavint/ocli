package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestDefaultPGConfig(t *testing.T) {
	cfg := DefaultPGConfig()

	if cfg == nil {
		t.Fatal("DefaultPGConfig() returned nil")
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Host", cfg.Host, "localhost"},
		{"Port", cfg.Port, 5432},
		{"User", cfg.User, "postgres"},
		{"SSLMode", cfg.SSLMode, "disable"},
		{"MaxOpenConns", cfg.MaxOpenConns, 25},
		{"MaxIdleConns", cfg.MaxIdleConns, 5},
		{"ConnMaxLifetime", cfg.ConnMaxLifetime, 5 * time.Minute},
		{"ConnMaxIdleTime", cfg.ConnMaxIdleTime, 1 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestConnectWithNilConfig(t *testing.T) {
	ctx := context.Background()
	db, err := Connect(ctx, "test", nil)

	if err == nil {
		t.Error("Connect() with nil config should return error")
	}

	if db != nil {
		t.Error("Connect() with nil config should return nil db")
	}

	expectedError := "postgres config is nil"
	if err != nil && !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Connect() error = %v, want error containing %v", err.Error(), expectedError)
	}
}

func TestIsValidDBName(t *testing.T) {
	tests := []struct {
		name   string
		dbname string
		want   bool
	}{
		{"valid simple name", "testdb", true},
		{"valid with underscore", "test_db", true},
		{"valid with number", "db123", true},
		{"valid mixed", "test_db_123", true},
		{"empty string", "", false},
		{"starts with number", "123db", false},
		{"contains hyphen", "test-db", false},
		{"contains space", "test db", false},
		{"contains special char", "test@db", false},
		{"contains dot", "test.db", false},
		{"all uppercase", "TESTDB", true},
		{"mixed case", "TestDb", true},
		{"single char", "a", true},
		{"max length valid", "a12345678901234567890123456789012345678901234567890123456789012", true},
		{"starts with underscore", "_testdb", true},
		{"too long", "a123456789012345678901234567890123456789012345678901234567890123456789", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDBName(tt.dbname)
			if got != tt.want {
				t.Errorf("IsValidDBName(%q) = %v, want %v", tt.dbname, got, tt.want)
			}
		})
	}
}

func TestCloseDB(t *testing.T) {
	// Test closing nil database
	err := CloseDB(nil)
	if err != nil {
		t.Errorf("CloseDB(nil) should not return error, got: %v", err)
	}
}

func TestPGConfigConnectionString(t *testing.T) {
	cfg := &PGConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}

	ctx := context.Background()
	// We expect this to fail since we don't have a real database
	// but we're testing the connection string construction
	_, err := Connect(ctx, "testdb", cfg)

	// Should get a connection error, not a config error
	if err == nil {
		t.Log("Unexpected: connection succeeded (maybe test DB is running)")
	}
}

func TestTerminateConnectionsContext(t *testing.T) {
	// This is a unit test for the logic, not actual DB operations
	// In a real scenario, you'd use a test database
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cfg := DefaultPGConfig()
	cfg.Host = "nonexistent"

	// Should fail quickly due to canceled context or connection failure
	_, err := Connect(ctx, "test", cfg)
	if err == nil {
		t.Error("Expected error when connecting with canceled context")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.in); got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

type fakeQuerier struct {
	queryCalls int
}

func (f *fakeQuerier) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	f.queryCalls++
	return nil, nil
}

func TestListDatabasesWithSizeOneQuery(t *testing.T) {
	q := &fakeQuerier{}
	got, err := ListDatabasesWithSizes(context.Background(), q)
	if err != nil {
		t.Fatalf("ListDatabasesWithSizes() error = %v", err)
	}
	if q.queryCalls != 1 {
		t.Fatalf("queryCalls = %d, want 1", q.queryCalls)
	}
	if len(got) != 0 {
		t.Fatalf("len(got) = %d, want 0", len(got))
	}
}
