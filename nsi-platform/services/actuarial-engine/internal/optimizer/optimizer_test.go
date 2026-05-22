package optimizer

import (
	"testing"
)

func TestFindParetoFrontier(t *testing.T) {
	schemes := []Scheme{
		{Name: "A", MonthlyCost: 500, ProjectedPension: 2000},
		{Name: "B", MonthlyCost: 1000, ProjectedPension: 3500},
		{Name: "C", MonthlyCost: 800, ProjectedPension: 3000},
		{Name: "D", MonthlyCost: 1500, ProjectedPension: 4000},
		{Name: "E", MonthlyCost: 500, ProjectedPension: 1500},
	}

	frontier := FindParetoFrontier(schemes)

	if len(frontier) == 0 {
		t.Fatal("expected non-empty frontier")
	}

	for _, f := range frontier {
		// A non-dominated scheme has NO other scheme that is
		// both cheaper AND has higher pension
		for _, s := range schemes {
			if s.Name == f.Name {
				continue
			}
			if s.MonthlyCost <= f.MonthlyCost && s.ProjectedPension >= f.ProjectedPension {
				if s.MonthlyCost < f.MonthlyCost || s.ProjectedPension > f.ProjectedPension {
					t.Errorf("%s is dominated by %s", f.Name, s.Name)
				}
			}
		}
	}
}

func TestFindParetoFrontierDominance(t *testing.T) {
	schemes := []Scheme{
		{Name: "A", MonthlyCost: 1000, ProjectedPension: 1000},
		{Name: "B", MonthlyCost: 500, ProjectedPension: 2000},
	}

	frontier := FindParetoFrontier(schemes)

	if len(frontier) != 1 {
		t.Fatalf("expected 1 frontier scheme, got %d", len(frontier))
	}
	if frontier[0].Name != "B" {
		t.Errorf("expected B, got %s", frontier[0].Name)
	}
}

func TestFindParetoFrontierEmpty(t *testing.T) {
	frontier := FindParetoFrontier([]Scheme{})
	if len(frontier) != 0 {
		t.Errorf("expected empty frontier, got %d", len(frontier))
	}
}

func TestFindParetoFrontierSingle(t *testing.T) {
	schemes := []Scheme{
		{Name: "A", MonthlyCost: 1000, ProjectedPension: 3000},
	}
	frontier := FindParetoFrontier(schemes)
	if len(frontier) != 1 {
		t.Errorf("expected 1 frontier scheme, got %d", len(frontier))
	}
	if frontier[0].Name != "A" {
		t.Errorf("expected A, got %s", frontier[0].Name)
	}
}

func TestFindParetoFrontierAllEqual(t *testing.T) {
	schemes := []Scheme{
		{Name: "A", MonthlyCost: 1000, ProjectedPension: 3000},
		{Name: "B", MonthlyCost: 1000, ProjectedPension: 3000},
	}
	frontier := FindParetoFrontier(schemes)
	if len(frontier) != 2 {
		t.Errorf("expected 2 frontier schemes for equal ones, got %d", len(frontier))
	}
}

func TestRankByEfficiency(t *testing.T) {
	schemes := []Scheme{
		{Name: "A", MonthlyCost: 500, ProjectedPension: 1500},
		{Name: "B", MonthlyCost: 1000, ProjectedPension: 4000},
		{Name: "C", MonthlyCost: 800, ProjectedPension: 2000},
	}

	ranked := RankByEfficiency(schemes)

	if len(ranked) != 3 {
		t.Fatalf("expected 3 ranked schemes, got %d", len(ranked))
	}
	// B: 4000/1000 = 4.0 (highest efficiency)
	if ranked[0].Name != "B" {
		t.Errorf("expected B as most efficient, got %s", ranked[0].Name)
	}
	// A: 1500/500 = 3.0
	if ranked[1].Name != "A" {
		t.Errorf("expected A as second, got %s", ranked[1].Name)
	}
	// C: 2000/800 = 2.5
	if ranked[2].Name != "C" {
		t.Errorf("expected C as third, got %s", ranked[2].Name)
	}
}

func TestRankByEfficiencyZeroCost(t *testing.T) {
	schemes := []Scheme{
		{Name: "A", MonthlyCost: 0, ProjectedPension: 1000},
		{Name: "B", MonthlyCost: 500, ProjectedPension: 2000},
	}

	ranked := RankByEfficiency(schemes)
	if len(ranked) != 2 {
		t.Fatalf("expected 2 ranked schemes, got %d", len(ranked))
	}
}

func TestRankByEfficiencyEmpty(t *testing.T) {
	ranked := RankByEfficiency([]Scheme{})
	if len(ranked) != 0 {
		t.Errorf("expected empty, got %d", len(ranked))
	}
}
