package elo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetExpectedScore(t *testing.T) {
	testcases := []struct {
		description string
		d           float64
		p           int32
		ra          int32
		rb          int32
		score       float64
	}{
		{
			description: "same rating",
			d:           400,
			p:           2,
			ra:          1000,
			rb:          1000,
			score:       0.5,
		},
		{
			description: "same rating, different d",
			d:           800,
			p:           2,
			ra:          1000,
			rb:          1000,
			score:       0.5,
		},
		{
			description: "ratingA is 25% higher",
			d:           400,
			p:           2,
			ra:          1250,
			rb:          1000,
			score:       0.81,
		},
		{
			description: "ratingB is 50% higher",
			d:           400,
			p:           2,
			ra:          1000,
			rb:          1500,
			score:       0.05,
		},
		{
			description: "ratingB is 50% higher with 3 decimals",
			d:           400,
			p:           3,
			ra:          1000,
			rb:          1500,
			score:       0.053,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.score, NewEloDefault().GetExpectedScore(tc.ra, tc.rb, tc.p))
		})
	}
}

func TestGetNewRatings(t *testing.T) {
	testCases := []struct {
		description string
		kCalculator KCalculator
		ra          int32
		rb          int32
		outcome     int32
		sCalculator SCalculator
		expErr      error
		expNewRa    int32
		expNewRb    int32
	}{
		{
			description: "negative Outcome",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     -1,
			sCalculator: NewSCalculatorLinear(),
			expErr:      ErrInvalidOutcome,
			expNewRa:    1500,
			expNewRb:    1500,
		},
		{
			description: "invalid Outcome",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     3,
			sCalculator: NewSCalculatorLinear(),
			expErr:      ErrInvalidOutcome,
			expNewRa:    1500,
			expNewRb:    1500,
		},
		{
			description: "same rating, playerA wins",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     0,
			sCalculator: NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    1516,
			expNewRb:    1484,
		},
		{
			description: "same rating, playerA wins with non-default k factor",
			kCalculator: NewKCalculatorConst(40),
			ra:          1500,
			rb:          1500,
			outcome:     0,
			sCalculator: NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    1520,
			expNewRb:    1480,
		},
		{
			description: "same rating, playerB wins",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     1,
			sCalculator: NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    1484,
			expNewRb:    1516,
		},
		{
			description: "same rating, Draw",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     2,
			sCalculator: NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    1500,
			expNewRb:    1500,
		},
		{
			description: "same rating, with mock sCalculator",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     2,
			sCalculator: &sCalculatorMock{},
			expErr:      nil,
			expNewRa:    1516,
			expNewRb:    1516,
		},
		{
			description: "ratingA is higher, playerA wins",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          2200,
			rb:          1900,
			outcome:     0,
			sCalculator: NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    2204,
			expNewRb:    1896,
		},
		{
			description: "ratingA is higher, playerB wins",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          2200,
			rb:          1900,
			outcome:     1,
			sCalculator: NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    2173,
			expNewRb:    1927,
		},
		{
			description: "ratingA is higher, Draw",
			kCalculator: NewKCalculatorConst(DefaultKFactor),
			ra:          2200,
			rb:          1900,
			outcome:     2,
			sCalculator: NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    2189,
			expNewRb:    1911,
		},
		{
			description: "ratingA is higher, a wins, USCF kFactor",
			kCalculator: NewKCalculatorUSCF(),
			ra:          2200,
			rb:          1900,
			outcome:     2,
			sCalculator: NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    2192,
			expNewRb:    1911,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			elo := NewElo(DefaultDValue, tc.sCalculator, tc.kCalculator)
			newRa, newRb, err := elo.GetNewRatings(tc.ra, tc.rb, tc.outcome)
			assert.ErrorIs(t, err, tc.expErr)
			assert.Equal(t, tc.expNewRa, newRa)
			assert.Equal(t, tc.expNewRb, newRb)
		})
	}
}

type sCalculatorMock struct{}

func (c *sCalculatorMock) getSValue(_ int32, _ Outcome) float64 {
	return 1
}
