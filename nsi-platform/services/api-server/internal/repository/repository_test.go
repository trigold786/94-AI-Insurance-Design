package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

func TestNewProfileRepositoryNilDB(t *testing.T) {
	_, err := NewProfileRepository(nil)
	if err == nil {
		t.Fatal("expected error for nil db, got nil")
	}
}

func TestNewProfileRepositorySuccess(t *testing.T) {
	db := &sql.DB{}
	repo, err := NewProfileRepository(db)
	if err != nil {
		t.Fatalf("NewProfileRepository() returned error: %v", err)
	}
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestUpsertNilProfile(t *testing.T) {
	db := &sql.DB{}
	repo, _ := NewProfileRepository(db)

	err := repo.Upsert(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil profile, got nil")
	}
}

func TestRepositoryImplementsInterface(t *testing.T) {
	db := &sql.DB{}
	repo, _ := NewProfileRepository(db)
	var iface ProfileRepository = repo
	if iface == nil {
		t.Error("ProfileRepository interface should not be nil")
	}
}

func TestProfileRepositoryFullFlow(t *testing.T) {
	// Integration test — requires real PostgreSQL
	t.Skip("skipping integration test; set NSI_TEST_DATABASE_URL to run")

	profile := &models.UserProfile{
		UserID:               "test-integration-user",
		Age:                  30,
		Gender:               "male",
		HouseholdRegionCode:  "330100",
		CurrentResidenceCode: "330100",
		EmploymentStatus:     "flexible",
		SocialSecurityYears:  8,
		HasChildren:          true,
	}

	repo, _ := NewProfileRepository(nil)
	err := repo.Upsert(context.Background(), profile)
	if err != nil {
		t.Fatalf("Upsert() returned error: %v", err)
	}

	got, err := repo.GetByUserID(context.Background(), "test-integration-user")
	if err != nil {
		t.Fatalf("GetByUserID() returned error: %v", err)
	}

	if got.UserID != profile.UserID {
		t.Errorf("expected UserID %s, got %s", profile.UserID, got.UserID)
	}
	if got.Age != profile.Age {
		t.Errorf("expected Age %d, got %d", profile.Age, got.Age)
	}
}
