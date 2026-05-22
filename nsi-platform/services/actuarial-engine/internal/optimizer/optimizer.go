package optimizer

import "sort"

type Scheme struct {
	Name             string
	MonthlyCost      float64
	ProjectedPension float64
}

func FindParetoFrontier(schemes []Scheme) []Scheme {
	if len(schemes) == 0 {
		return []Scheme{}
	}

	var frontier []Scheme
	for _, s := range schemes {
		dominated := false
		for _, o := range schemes {
			if s.Name == o.Name {
				continue
			}
			if o.MonthlyCost <= s.MonthlyCost && o.ProjectedPension >= s.ProjectedPension {
				if o.MonthlyCost < s.MonthlyCost || o.ProjectedPension > s.ProjectedPension {
					dominated = true
					break
				}
			}
		}
		if !dominated {
			frontier = append(frontier, s)
		}
	}

	return frontier
}

func RankByEfficiency(schemes []Scheme) []Scheme {
	if len(schemes) == 0 {
		return []Scheme{}
	}

	sorted := make([]Scheme, len(schemes))
	copy(sorted, schemes)

	sort.Slice(sorted, func(i, j int) bool {
		ei := efficiency(sorted[i])
		ej := efficiency(sorted[j])
		return ei > ej
	})

	return sorted
}

func efficiency(s Scheme) float64 {
	if s.MonthlyCost <= 0 {
		return s.ProjectedPension
	}
	return s.ProjectedPension / s.MonthlyCost
}
