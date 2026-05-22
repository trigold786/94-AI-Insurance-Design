package cashflow

import (
	"math"
	"testing"
)

func TestProjectSingleYear(t *testing.T) {
	items := Project(CashflowInput{
		Years:            1,
		AnnualPayment:    12000,
		AnnualSubsidy:    2400,
		InitialBalance:   0,
		SalaryGrowthRate: 0,
		FundReturnRate:   0,
	})

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Year != 1 {
		t.Errorf("expected year 1, got %d", items[0].Year)
	}
	if items[0].Payment != 12000 {
		t.Errorf("expected payment 12000, got %.2f", items[0].Payment)
	}
	if items[0].Subsidy != 2400 {
		t.Errorf("expected subsidy 2400, got %.2f", items[0].Subsidy)
	}
	if items[0].Balance != 14400 {
		t.Errorf("expected balance 14400, got %.2f", items[0].Balance)
	}
}

func TestProjectMultipleYears(t *testing.T) {
	items := Project(CashflowInput{
		Years:            3,
		AnnualPayment:    12000,
		AnnualSubsidy:    2400,
		InitialBalance:   0,
		SalaryGrowthRate: 0,
		FundReturnRate:   0,
	})

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Each year: payment stays 12000, subsidy stays 2400, balance accumulates
	expectedBalances := []float64{14400, 28800, 43200}
	for i, exp := range expectedBalances {
		if items[i].Balance != exp {
			t.Errorf("year %d: expected balance %.2f, got %.2f", i+1, exp, items[i].Balance)
		}
	}
}

func almostEqual(a, b, eps float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < eps
}

func TestProjectWithSalaryGrowth(t *testing.T) {
	items := Project(CashflowInput{
		Years:            3,
		AnnualPayment:    10000,
		AnnualSubsidy:    2000,
		InitialBalance:   0,
		SalaryGrowthRate: 0.10,
		FundReturnRate:   0,
	})

	// Year 1: pay=10000, sub=2000, bal=12000
	// Year 2: pay=11000, sub=2200, bal=25200
	// Year 3: pay=12100, sub=2420, bal=39720
	if !almostEqual(items[0].Payment, 10000, 0.01) {
		t.Errorf("year 1 payment: expected 10000, got %.4f", items[0].Payment)
	}
	if !almostEqual(items[1].Payment, 11000, 0.01) {
		t.Errorf("year 2 payment: expected 11000, got %.4f", items[1].Payment)
	}
	if !almostEqual(items[2].Payment, 12100, 0.01) {
		t.Errorf("year 3 payment: expected 12100, got %.4f", items[2].Payment)
	}
	if !almostEqual(items[1].Subsidy, 2200, 0.01) {
		t.Errorf("year 2 subsidy: expected 2200, got %.4f", items[1].Subsidy)
	}
	expectedBalances := []float64{12000, 25200, 39720}
	for i, exp := range expectedBalances {
		if !almostEqual(items[i].Balance, exp, 0.01) {
			t.Errorf("year %d balance: expected %.2f, got %.4f", i+1, exp, items[i].Balance)
		}
	}
}

func TestProjectWithFundReturn(t *testing.T) {
	items := Project(CashflowInput{
		Years:            2,
		AnnualPayment:    12000,
		AnnualSubsidy:    2400,
		InitialBalance:   0,
		SalaryGrowthRate: 0,
		FundReturnRate:   0.05,
	})

	// Year 1: pay=12000, sub=2400, bal=14400, return=0 (first year after contribution)
	// Year 2: pay=12000, sub=2400, bal = 14400*1.05 + 14400 = 29520
	if !almostEqual(items[1].Balance, 29520, 0.01) {
		t.Errorf("year 2 balance: expected 29520, got %.2f", items[1].Balance)
	}
}

func TestProjectWithInitialBalance(t *testing.T) {
	items := Project(CashflowInput{
		Years:            1,
		AnnualPayment:    12000,
		AnnualSubsidy:    2400,
		InitialBalance:   50000,
		SalaryGrowthRate: 0,
		FundReturnRate:   0,
	})

	if items[0].Balance != 64400 {
		t.Errorf("expected balance 64400, got %.2f", items[0].Balance)
	}
}

func TestProjectZeroYears(t *testing.T) {
	items := Project(CashflowInput{
		Years:            0,
		AnnualPayment:    12000,
		AnnualSubsidy:    2400,
		InitialBalance:   0,
		SalaryGrowthRate: 0,
		FundReturnRate:   0,
	})

	if len(items) != 0 {
		t.Errorf("expected 0 items for 0 years, got %d", len(items))
	}
}

func TestProjectNegativePayment(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not panic on negative payment: %v", r)
		}
	}()

	items := Project(CashflowInput{
		Years:            1,
		AnnualPayment:    -1000,
		AnnualSubsidy:    0,
		InitialBalance:   0,
		SalaryGrowthRate: 0,
		FundReturnRate:   0,
	})

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Payment != -1000 {
		t.Errorf("expected payment -1000, got %.2f", items[0].Payment)
	}
}

func TestCashFlowItemFields(t *testing.T) {
	item := CashFlowItem{Year: 5, Payment: 1000, Subsidy: 200, Balance: 6000}
	if item.Year != 5 || item.Payment != 1000 || item.Subsidy != 200 || item.Balance != 6000 {
		t.Error("CashFlowItem fields not set correctly")
	}
}

func TestProjectRatesWithinBounds(t *testing.T) {
	// Test with moderately high realistic rates
	items := Project(CashflowInput{
		Years:            30,
		AnnualPayment:    24000,
		AnnualSubsidy:    5000,
		InitialBalance:   0,
		SalaryGrowthRate: 0.05,
		FundReturnRate:   0.03,
	})

	if len(items) != 30 {
		t.Fatalf("expected 30 items, got %d", len(items))
	}

	// Balance should grow over time
	for i := 1; i < len(items); i++ {
		if items[i].Balance < items[i-1].Balance {
			t.Errorf("year %d balance %.2f < year %d balance %.2f",
				i+1, items[i].Balance, i, items[i-1].Balance)
		}
	}

	// Final balance should be positive and significant
	if items[29].Balance <= 0 {
		t.Errorf("expected positive final balance, got %.2f", items[29].Balance)
	}
	if math.IsNaN(items[29].Balance) || math.IsInf(items[29].Balance, 0) {
		t.Errorf("balance should be finite, got %.2f", items[29].Balance)
	}
}
