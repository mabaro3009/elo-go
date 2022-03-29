package elo

type KCalculatorConst struct {
	k float64
}

func NewKCalculatorConst(k float64) *KCalculatorConst {
	return &KCalculatorConst{k: k}
}

func (c *KCalculatorConst) getKFactor(_ int32) float64 {
	return c.k
}

type KCalculatorUSCF struct{}

func NewKCalculatorUSCF() *KCalculatorUSCF {
	return &KCalculatorUSCF{}
}

func (c *KCalculatorUSCF) getKFactor(rating int32) float64 {
	switch {
	case rating < 2100:
		return 32
	case rating >= 2100 && rating <= 2400:
		return 24
	default:
		return 16
	}
}

type KCalculatorFIDESimplified struct{}

func NewKCalculatorFIDESimplified() *KCalculatorFIDESimplified {
	return &KCalculatorFIDESimplified{}
}

func (c *KCalculatorFIDESimplified) getKFactor(rating int32) float64 {
	switch {
	case rating < 2300:
		return 40
	case rating >= 2300 && rating <= 2400:
		return 20
	default:
		return 10
	}
}
