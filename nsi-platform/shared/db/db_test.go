package db

import (
	"os"
	"testing"
)

func TestConnectSuccess(t *testing.T) {
	dsn := os.Getenv("NSI_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NSI_TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := Connect(dsn)
	if err != nil {
		t.Fatalf("Connect() returned error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping() returned error: %v", err)
	}
}

func TestConnectInvalidDSN(t *testing.T) {
	db, err := Connect("postgres://invalid:invalid@localhost:1/test?sslmode=disable")
	if err == nil {
		db.Close()
		t.Fatal("expected error for invalid DSN, got nil")
	}
}

func TestConnectEmptyDSN(t *testing.T) {
	db, err := Connect("")
	if err == nil {
		db.Close()
		t.Fatal("expected error for empty DSN, got nil")
	}
}
