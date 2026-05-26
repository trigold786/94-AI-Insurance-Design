package city

import "fmt"

type Params struct {
	Name               string
	AvgSalary          float64
	ContributionBaseMin float64
	ContributionBaseMax float64
	PensionEmployeeRate  float64
	PensionEmployerRate  float64
	MedicalEmployeeRate  float64
	MedicalEmployerRate  float64
	UnemploymentEmployeeRate float64
	UnemploymentEmployerRate float64
	WorkerCompEmployerRate  float64
	MaternityEmployerRate   float64
	MalePensionAge      int
	FemalePensionAge    int
}

func (p *Params) CalcMonthlyCost(salary float64) float64 {
	employeeTotal := p.PensionEmployeeRate + p.MedicalEmployeeRate + p.UnemploymentEmployeeRate
	employerTotal := p.PensionEmployerRate + p.MedicalEmployerRate + p.UnemploymentEmployerRate +
		p.WorkerCompEmployerRate + p.MaternityEmployerRate
	return salary * (employeeTotal + employerTotal)
}

func (p *Params) TotalEmployerRate() float64 {
	return p.PensionEmployerRate + p.MedicalEmployerRate + p.UnemploymentEmployerRate +
		p.WorkerCompEmployerRate + p.MaternityEmployerRate
}

var builtinCities = map[string]Params{
	"310000": {
		Name:              "上海",
		AvgSalary:          12383,
		ContributionBaseMin: 7380,
		ContributionBaseMax: 36948,
		PensionEmployeeRate:  0.08,
		PensionEmployerRate:  0.16,
		MedicalEmployeeRate:  0.02,
		MedicalEmployerRate:  0.09,
		UnemploymentEmployeeRate: 0.005,
		UnemploymentEmployerRate: 0.005,
		WorkerCompEmployerRate:  0.005,
		MaternityEmployerRate:   0.01,
		MalePensionAge:   60,
		FemalePensionAge: 50,
	},
	"110000": {
		Name:              "北京",
		AvgSalary:          15764,
		ContributionBaseMin: 9460,
		ContributionBaseMax: 47292,
		PensionEmployeeRate:  0.08,
		PensionEmployerRate:  0.16,
		MedicalEmployeeRate:  0.02,
		MedicalEmployerRate:  0.098,
		UnemploymentEmployeeRate: 0.005,
		UnemploymentEmployerRate: 0.005,
		WorkerCompEmployerRate:  0.004,
		MaternityEmployerRate:   0.008,
		MalePensionAge:   60,
		FemalePensionAge: 50,
	},
	"440300": {
		Name:              "深圳",
		AvgSalary:          14530,
		ContributionBaseMin: 6525,
		ContributionBaseMax: 43590,
		PensionEmployeeRate:  0.08,
		PensionEmployerRate:  0.16,
		MedicalEmployeeRate:  0.02,
		MedicalEmployerRate:  0.062,
		UnemploymentEmployeeRate: 0.003,
		UnemploymentEmployerRate: 0.007,
		WorkerCompEmployerRate:  0.0014,
		MaternityEmployerRate:   0.0045,
		MalePensionAge:   60,
		FemalePensionAge: 50,
	},
	"440100": {
		Name:              "广州",
		AvgSalary:          13795,
		ContributionBaseMin: 8280,
		ContributionBaseMax: 41385,
		PensionEmployeeRate:  0.08,
		PensionEmployerRate:  0.16,
		MedicalEmployeeRate:  0.02,
		MedicalEmployerRate:  0.0685,
		UnemploymentEmployeeRate: 0.002,
		UnemploymentEmployerRate: 0.008,
		WorkerCompEmployerRate:  0.0016,
		MaternityEmployerRate:   0.0085,
		MalePensionAge:   60,
		FemalePensionAge: 50,
	},
	"330100": {
		Name:              "杭州",
		AvgSalary:          9625,
		ContributionBaseMin: 5850,
		ContributionBaseMax: 28875,
		PensionEmployeeRate:  0.08,
		PensionEmployerRate:  0.15,
		MedicalEmployeeRate:  0.02,
		MedicalEmployerRate:  0.095,
		UnemploymentEmployeeRate: 0.005,
		UnemploymentEmployerRate: 0.005,
		WorkerCompEmployerRate:  0.004,
		MaternityEmployerRate:   0.01,
		MalePensionAge:   60,
		FemalePensionAge: 50,
	},
}

func GetParams(cityCode string) (*Params, error) {
	if cityCode == "" {
		return nil, fmt.Errorf("city code cannot be empty")
	}
	p, ok := builtinCities[cityCode]
	if !ok {
		return nil, fmt.Errorf("unknown city code: %s", cityCode)
	}
	cp := p
	return &cp, nil
}
