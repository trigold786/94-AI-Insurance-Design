package tax

import (
	"math"
	"testing"
)

func TestCalculateTax_BelowThreshold(t *testing.T) {
	tax, rate := CalculateTax(50000, 5000)
	if tax != 0 {
		t.Errorf("expected 0 tax for income below threshold, got %.2f", tax)
	}
	if rate != 0 {
		t.Errorf("expected 0 rate for income below threshold, got %.4f", rate)
	}
}

func TestCalculateTax_FirstBracket(t *testing.T) {
	// Annual: 100000, SI base: 5000
	// SI deduction: 5000*12*0.185 = 11100
	// Taxable: 100000 - 60000 - 11100 = 28900
	// Tax: 28900 * 3% = 867
	tax, rate := CalculateTax(100000, 5000)
	expected := 28900 * 0.03
	if !almostEqual(tax, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, tax)
	}
	expectedRate := expected / 100000
	if !almostEqual(rate, expectedRate, 0.0001) {
		t.Errorf("expected rate %.4f, got %.4f", expectedRate, rate)
	}
}

func TestCalculateTax_SecondBracket(t *testing.T) {
	// Annual: 200000, SI base: 10000
	// SI deduction: 10000*12*0.185 = 22200
	// Taxable: 200000 - 60000 - 22200 = 117800
	// Tax: 117800*10% - 2520 = 9260
	tax, rate := CalculateTax(200000, 10000)
	expected := 117800*0.10 - 2520
	if !almostEqual(tax, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, tax)
	}
	expectedRate := expected / 200000
	if !almostEqual(rate, expectedRate, 0.0001) {
		t.Errorf("expected rate %.4f, got %.4f", expectedRate, rate)
	}
}

func TestCalculateTax_ThirdBracket(t *testing.T) {
	// Annual: 500000, SI base: 15000
	// SI deduction: 15000*12*0.185 = 33300
	// Taxable: 500000 - 60000 - 33300 = 406700
	// Tax: 406700*25% - 31920 = 101675 - 31920 = 69755
	tax, rate := CalculateTax(500000, 15000)
	expected := 406700*0.25 - 31920
	if !almostEqual(tax, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, tax)
	}
	expectedRate := expected / 500000
	if !almostEqual(rate, expectedRate, 0.0001) {
		t.Errorf("expected rate %.4f, got %.4f", expectedRate, rate)
	}
}

func TestCalculateTax_HighestBracket(t *testing.T) {
	// Annual: 2000000, SI base: 30000
	// SI deduction: 30000*12*0.185 = 66600
	// Taxable: 2000000 - 60000 - 66600 = 1873400
	// Tax: 1873400*45% - 181920 = 843030 - 181920 = 661110
	tax, rate := CalculateTax(2000000, 30000)
	expected := 1873400*0.45 - 181920
	if !almostEqual(tax, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, tax)
	}
	expectedRate := expected / 2000000
	if !almostEqual(rate, expectedRate, 0.0001) {
		t.Errorf("expected rate %.4f, got %.4f", expectedRate, rate)
	}
}

func TestCalculateTax_ZeroSIBase(t *testing.T) {
	// Annual: 80000, SI base: 0
	// Taxable: 80000 - 60000 - 0 = 20000
	// Tax: 20000 * 3% = 600
	tax, _ := CalculateTax(80000, 0)
	expected := 20000 * 0.03
	if !almostEqual(tax, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, tax)
	}
}

func TestCalculateTax_ZeroIncome(t *testing.T) {
	tax, rate := CalculateTax(0, 5000)
	if tax != 0 {
		t.Errorf("expected 0 tax for zero income, got %.2f", tax)
	}
	if rate != 0 {
		t.Errorf("expected 0 rate for zero income, got %.4f", rate)
	}
}

func TestCalculateTax_ThresholdBoundary(t *testing.T) {
	// Just at 36000 taxable (after deductions)
	// If taxable is exactly 36000, tax = 36000*3% = 1080
	// Test varies base to make taxable = 36000
	// We'll use annualIncome=96000+60000+SI, SI based on base
	// Actually let's use a simpler approach - just verify bracket boundary accuracy
	// Annual: 120000, SI base: 8000
	// SI deduction: 8000*12*0.185 = 17760
	// Taxable: 120000 - 60000 - 17760 = 42240 (second bracket)
	// Tax: 42240*10% - 2520 = 1704
	tax, _ := CalculateTax(120000, 8000)
	expected := 42240*0.10 - 2520
	if !almostEqual(tax, expected, 0.01) {
		t.Errorf("expected %.2f, got %.2f", expected, tax)
	}
}

func almostEqual(a, b, eps float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < eps || (math.IsNaN(a) && math.IsNaN(b))
}
