package handler

type CityInfo struct {
	Name             string
	AvgSalary        float64
	MalePensionAge   int
	FemalePensionAge int
}

var builtinCities = map[string]CityInfo{
	"310000": {Name: "上海", AvgSalary: 12383, MalePensionAge: 60, FemalePensionAge: 50},
	"110000": {Name: "北京", AvgSalary: 15764, MalePensionAge: 60, FemalePensionAge: 50},
	"440300": {Name: "深圳", AvgSalary: 14530, MalePensionAge: 60, FemalePensionAge: 50},
	"440100": {Name: "广州", AvgSalary: 13795, MalePensionAge: 60, FemalePensionAge: 50},
	"330100": {Name: "杭州", AvgSalary: 9625, MalePensionAge: 60, FemalePensionAge: 50},
}

func GetCityInfo(cityCode string) *CityInfo {
	c, ok := builtinCities[cityCode]
	if !ok {
		return nil
	}
	return &c
}
