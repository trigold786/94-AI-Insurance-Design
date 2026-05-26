package tax

// CalculateTax computes Chinese personal income tax (个税) using cumulative withholding method.
// annualTaxableIncome: total annual pre-tax income
// socialInsuranceBase: monthly base for social insurance deductions
// Returns (annualTax, effectiveRate)
func CalculateTax(annualTaxableIncome float64, socialInsuranceBase float64) (float64, float64) {
	if annualTaxableIncome <= 0 {
		return 0, 0
	}

	// Standard deduction: 5000/month
	standardDeduction := 60000.0

	// Social insurance deductions: 养老8% + 医疗2% + 失业0.5% + 住房8% = 18.5%
	insuranceRate := 0.185
	socialInsuranceDeduction := socialInsuranceBase * 12 * insuranceRate

	taxableIncome := annualTaxableIncome - standardDeduction - socialInsuranceDeduction
	if taxableIncome <= 0 {
		return 0, 0
	}

	type bracket struct {
		upper      float64
		rate       float64
		quickDed   float64
	}

	brackets := []bracket{
		{36000, 0.03, 0},
		{144000, 0.10, 2520},
		{300000, 0.20, 16920},
		{420000, 0.25, 31920},
		{660000, 0.30, 52920},
		{960000, 0.35, 85920},
		{0, 0.45, 181920},
	}

	var tax float64
	for _, b := range brackets {
		if b.upper == 0 || taxableIncome <= b.upper {
			tax = taxableIncome*b.rate - b.quickDed
			break
		}
	}

	if tax < 0 {
		tax = 0
	}

	effectiveRate := tax / annualTaxableIncome

	return tax, effectiveRate
}
