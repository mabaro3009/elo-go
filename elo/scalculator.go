package elo

type SCalculatorLinear struct{}

func NewSCalculatorLinear() *SCalculatorLinear {
	return &SCalculatorLinear{}
}

// getSValue returns a linear score for s.
// o: 0 for Win, 1 for Loss, 2 for Draw
func (c *SCalculatorLinear) getSValue(n int32, o Outcome) float64 {
	switch o {
	case Win:
		return 1
	case Loss:
		return 0
	default:
		return 1 / float64(n)
	}
}
