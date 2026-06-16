package optimizer

import (
	"math"
	"math/rand"
	"sort"
)

type Scheme struct {
	Name             string
	MonthlyCost      float64
	ProjectedPension float64
	EquityScore      float64
}

type nsgaIndividual struct {
	scheme    Scheme
	rank      int
	crowding  float64
}

func NSGAII(schemes []Scheme, populationSize, generations int) []Scheme {
	if len(schemes) == 0 {
		return []Scheme{}
	}
	if len(schemes) <= 5 {
		return FilterParetoOptimal(schemes, 0)
	}

	pop := make([]nsgaIndividual, len(schemes))
	for i, s := range schemes {
		pop[i] = nsgaIndividual{scheme: s}
	}

	for gen := 0; gen < generations; gen++ {
		pop = nonDominatedSort(pop)
		pop = calculateCrowdingDistance(pop)

		if len(pop) > populationSize {
			sort.Slice(pop, func(i, j int) bool {
				if pop[i].rank != pop[j].rank {
					return pop[i].rank < pop[j].rank
				}
				return pop[i].crowding > pop[j].crowding
			})
			pop = pop[:populationSize]
		}

		if gen < generations-1 {
			offspring := make([]nsgaIndividual, 0, len(pop))
			for i := 0; i < len(pop); i += 2 {
				p1 := tournamentSelect(pop)
				p2 := tournamentSelect(pop)
				c1, c2 := crossover(p1, p2)
				c1 = mutate(c1)
				c2 = mutate(c2)
				offspring = append(offspring, c1, c2)
			}
			pop = append(pop, offspring...)
		}
	}

	pop = nonDominatedSort(pop)
	rank1 := []Scheme{}
	for _, ind := range pop {
		if ind.rank == 1 {
			rank1 = append(rank1, ind.scheme)
		}
	}
	if len(rank1) > 5 {
		rank1 = rank1[:5]
	}
	if len(rank1) == 0 {
		return FilterParetoOptimal(schemes, 0)
	}
	return rank1
}

func nonDominatedSort(pop []nsgaIndividual) []nsgaIndividual {
	for i := range pop {
		pop[i].rank = 1
	}
	for i := range pop {
		for j := range pop {
			if i == j {
				continue
			}
			if dominates(pop[j].scheme, pop[i].scheme) {
				pop[i].rank++
			}
		}
	}
	sort.Slice(pop, func(i, j int) bool {
		return pop[i].rank < pop[j].rank
	})
	return pop
}

func calculateCrowdingDistance(pop []nsgaIndividual) []nsgaIndividual {
	if len(pop) <= 2 {
		for i := range pop {
			pop[i].crowding = math.Inf(1)
		}
		return pop
	}
	for i := range pop {
		pop[i].crowding = 0
	}
	objectives := []func(Scheme) float64{
		func(s Scheme) float64 { return -s.MonthlyCost },
		func(s Scheme) float64 { return s.ProjectedPension },
		func(s Scheme) float64 { return s.EquityScore },
	}
	for _, obj := range objectives {
		sort.Slice(pop, func(i, j int) bool {
			return obj(pop[i].scheme) < obj(pop[j].scheme)
		})
		pop[0].crowding = math.Inf(1)
		pop[len(pop)-1].crowding = math.Inf(1)
		minVal := obj(pop[0].scheme)
		maxVal := obj(pop[len(pop)-1].scheme)
		rangeVal := maxVal - minVal
		if rangeVal == 0 {
			rangeVal = 1
		}
		for i := 1; i < len(pop)-1; i++ {
			pop[i].crowding += (obj(pop[i+1].scheme) - obj(pop[i-1].scheme)) / rangeVal
		}
	}
	return pop
}

func tournamentSelect(pop []nsgaIndividual) nsgaIndividual {
	i := rand.Intn(len(pop))
	j := rand.Intn(len(pop))
	if pop[i].rank < pop[j].rank || (pop[i].rank == pop[j].rank && pop[i].crowding > pop[j].crowding) {
		return pop[i]
	}
	return pop[j]
}

func crossover(a, b nsgaIndividual) (nsgaIndividual, nsgaIndividual) {
	w := rand.Float64()
	c1 := nsgaIndividual{scheme: blendScheme(a.scheme, b.scheme, w)}
	c2 := nsgaIndividual{scheme: blendScheme(b.scheme, a.scheme, w)}
	return c1, c2
}

func blendScheme(a, b Scheme, w float64) Scheme {
	return Scheme{
		Name:             a.Name,
		MonthlyCost:      a.MonthlyCost*w + b.MonthlyCost*(1-w),
		ProjectedPension: a.ProjectedPension*w + b.ProjectedPension*(1-w),
		EquityScore:      a.EquityScore*w + b.EquityScore*(1-w),
	}
}

func mutate(ind nsgaIndividual) nsgaIndividual {
	if rand.Float64() < 0.1 {
		ind.scheme.MonthlyCost *= (0.95 + rand.Float64()*0.1)
		ind.scheme.ProjectedPension *= (0.95 + rand.Float64()*0.1)
		ind.scheme.EquityScore *= (0.95 + rand.Float64()*0.1)
	}
	return ind
}

func dominates(a, b Scheme) bool {
	costBetter := a.MonthlyCost <= b.MonthlyCost
	pensionBetter := a.ProjectedPension >= b.ProjectedPension
	equityBetter := a.EquityScore >= b.EquityScore
	if !costBetter || !pensionBetter || !equityBetter {
		return false
	}
	return a.MonthlyCost < b.MonthlyCost || a.ProjectedPension > b.ProjectedPension || a.EquityScore > b.EquityScore
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
			if dominates(o, s) {
				dominated = true
				break
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
		return efficiency(sorted[i]) > efficiency(sorted[j])
	})
	return sorted
}

func FilterParetoOptimal(schemes []Scheme, efficiencyThreshold float64) []Scheme {
	frontier := FindParetoFrontier(schemes)
	if efficiencyThreshold <= 0 {
		return frontier
	}
	var filtered []Scheme
	for _, s := range frontier {
		if efficiency(s) >= efficiencyThreshold {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func efficiency(s Scheme) float64 {
	if s.MonthlyCost <= 0 {
		return s.ProjectedPension
	}
	return s.ProjectedPension / s.MonthlyCost
}

func SelectTopN(schemes []Scheme, n int) []Scheme {
	if len(schemes) <= n {
		return schemes
	}
	ranked := RankByEfficiency(schemes)
	if len(ranked) > n {
		ranked = ranked[:n]
	}
	return ranked
}
