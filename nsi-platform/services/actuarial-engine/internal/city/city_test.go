package city

import (
	"testing"
)

func TestGetParamsSuccess(t *testing.T) {
	cases := []struct {
		code   string
		name   string
	}{
		{"310000", "上海"},
		{"110000", "北京"},
		{"440300", "深圳"},
		{"440100", "广州"},
		{"330100", "杭州"},
	}
	for _, c := range cases {
		p, err := GetParams(c.code)
		if err != nil {
			t.Errorf("GetParams(%s): unexpected error: %v", c.code, err)
		}
		if p.AvgSalary <= 0 {
			t.Errorf("GetParams(%s): expected AvgSalary > 0, got %f", c.code, p.AvgSalary)
		}
		if p.PensionEmployeeRate <= 0 || p.PensionEmployeeRate > 1 {
			t.Errorf("GetParams(%s): PensionEmployeeRate out of range: %f", c.code, p.PensionEmployeeRate)
		}
		if p.PensionEmployerRate <= 0 || p.PensionEmployerRate > 1 {
			t.Errorf("GetParams(%s): PensionEmployerRate out of range: %f", c.code, p.PensionEmployerRate)
		}
		if p.ContributionBaseMin > p.ContributionBaseMax {
			t.Errorf("GetParams(%s): ContributionBaseMin(%f) > ContributionBaseMax(%f)", c.code, p.ContributionBaseMin, p.ContributionBaseMax)
		}
		if p.Name != c.name {
			t.Errorf("GetParams(%s): expected Name %s, got %s", c.code, c.name, p.Name)
		}
	}
}

func TestGetParamsNotFound(t *testing.T) {
	_, err := GetParams("999999")
	if err == nil {
		t.Fatal("expected error for unknown city code, got nil")
	}
}

func TestGetParamsEmptyCode(t *testing.T) {
	_, err := GetParams("")
	if err == nil {
		t.Fatal("expected error for empty code, got nil")
	}
}

func TestCalcMonthlyCost(t *testing.T) {
	sh, _ := GetParams("310000")
	cost := sh.CalcMonthlyCost(12383)
	if cost <= 0 {
		t.Errorf("expected positive monthly cost, got %f", cost)
	}
}

func TestTotalEmployerRate(t *testing.T) {
	sh, _ := GetParams("310000")
	total := sh.TotalEmployerRate()
	if total <= 0 || total > 1 {
		t.Errorf("TotalEmployerRate out of range: %f", total)
	}
}

func TestAllCitiesHaveValidAges(t *testing.T) {
	for code := range builtinCities {
		p, _ := GetParams(code)
		if p.MalePensionAge < 55 || p.MalePensionAge > 70 {
			t.Errorf("%s: MalePensionAge out of range: %d", code, p.MalePensionAge)
		}
		if p.FemalePensionAge < 45 || p.FemalePensionAge > 65 {
			t.Errorf("%s: FemalePensionAge out of range: %d", code, p.FemalePensionAge)
		}
	}
}
