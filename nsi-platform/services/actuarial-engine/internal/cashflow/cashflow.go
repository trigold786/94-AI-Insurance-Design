package cashflow

type CashFlowItem struct {
	Year    int     `json:"year"`
	Payment float64 `json:"payment"`
	Subsidy float64 `json:"subsidy"`
	Balance float64 `json:"balance"`
}

type CashflowInput struct {
	Years            int
	AnnualPayment    float64
	AnnualSubsidy    float64
	InitialBalance   float64
	SalaryGrowthRate float64
	FundReturnRate   float64
}

func Project(input CashflowInput) []CashFlowItem {
	if input.Years <= 0 {
		return []CashFlowItem{}
	}

	items := make([]CashFlowItem, 0, input.Years)
	balance := input.InitialBalance
	payment := input.AnnualPayment
	subsidy := input.AnnualSubsidy

	for year := 1; year <= input.Years; year++ {
		balance = balance * (1 + input.FundReturnRate)
		balance += payment + subsidy

		items = append(items, CashFlowItem{
			Year:    year,
			Payment: payment,
			Subsidy: subsidy,
			Balance: balance,
		})

		payment *= (1 + input.SalaryGrowthRate)
		subsidy *= (1 + input.SalaryGrowthRate)
	}

	return items
}
