package repository

import (
	"database/sql"
	"testing"
)

func TestNewPolicyRepositoryNilDB(t *testing.T) {
	_, err := NewPolicyRepository(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
}

func TestNewPolicyRepositorySuccess(t *testing.T) {
	db := &sql.DB{}
	repo, err := NewPolicyRepository(db)
	if err != nil {
		t.Fatalf("NewPolicyRepository() returned error: %v", err)
	}
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestPolicyRepositoryImplementsInterface(t *testing.T) {
	db := &sql.DB{}
	repo, _ := NewPolicyRepository(db)
	var iface PolicyRepository = repo
	if iface == nil {
		t.Error("PolicyRepository interface should not be nil")
	}
}

func TestPolicyFilterValidation(t *testing.T) {
	if (&PolicyFilter{}).Limit != 0 {
		t.Error("expected zero value Limit")
	}
	if (&PolicyFilter{RegionCode: "310000"}).RegionCode != "310000" {
		t.Error("RegionCode not set correctly")
	}
}
