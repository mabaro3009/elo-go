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
		k           KCalculator
		ra          int32
		rb          int32
		outcome     int32
		s           SCalculator
		expErr      error
		expNewRa    int32
		expNewRb    int32
	}{
		{
			description: "negative Outcome",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     -1,
			s:           NewSCalculatorLinear(),
			expErr:      ErrInvalidOutcome,
			expNewRa:    1500,
			expNewRb:    1500,
		},
		{
			description: "invalid Outcome",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     3,
			s:           NewSCalculatorLinear(),
			expErr:      ErrInvalidOutcome,
			expNewRa:    1500,
			expNewRb:    1500,
		},
		{
			description: "same rating, playerA wins",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     0,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    1516,
			expNewRb:    1484,
		},
		{
			description: "same rating, playerA wins with non-default k factor",
			k:           NewKCalculatorConst(40),
			ra:          1500,
			rb:          1500,
			outcome:     0,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    1520,
			expNewRb:    1480,
		},
		{
			description: "same rating, playerB wins",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     1,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    1484,
			expNewRb:    1516,
		},
		{
			description: "same rating, Draw",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     2,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    1500,
			expNewRb:    1500,
		},
		{
			description: "same rating, with mock s",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          1500,
			rb:          1500,
			outcome:     2,
			s:           &sCalculatorMock{},
			expErr:      nil,
			expNewRa:    1516,
			expNewRb:    1516,
		},
		{
			description: "ratingA is higher, playerA wins",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          2200,
			rb:          1900,
			outcome:     0,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    2204,
			expNewRb:    1896,
		},
		{
			description: "ratingA is higher, playerB wins",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          2200,
			rb:          1900,
			outcome:     1,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    2173,
			expNewRb:    1927,
		},
		{
			description: "ratingA is higher, Draw",
			k:           NewKCalculatorConst(DefaultKFactor),
			ra:          2200,
			rb:          1900,
			outcome:     2,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    2189,
			expNewRb:    1911,
		},
		{
			description: "ratingA is higher, a wins, USCF kFactor",
			k:           NewKCalculatorUSCF(),
			ra:          2200,
			rb:          1900,
			outcome:     2,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewRa:    2192,
			expNewRb:    1911,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			elo := NewElo(DefaultDValue, tc.s, tc.k)
			newRa, newRb, err := elo.GetNewRatings(tc.ra, tc.rb, tc.outcome)
			assert.ErrorIs(t, err, tc.expErr)
			assert.Equal(t, tc.expNewRa, newRa)
			assert.Equal(t, tc.expNewRb, newRb)
		})
	}
}

func TestGetNewRatingsTeams(t *testing.T) {
	testCases := []struct {
		description string
		ratingsA    []int32
		ratingsB    []int32
		out         int32
		expError    error
		expRatingsA []int32
		expRatingsB []int32
	}{
		{
			description: "team length missmatch.",
			ratingsA:    []int32{1500, 1500, 1500},
			ratingsB:    []int32{1500, 1500},
			out:         0,
			expError:    ErrTeamLenMissmatch,
			expRatingsA: []int32{1500, 1500, 1500},
			expRatingsB: []int32{1500, 1500},
		},
		{
			description: "invalid outcome. Negative.",
			ratingsA:    []int32{1500, 1500},
			ratingsB:    []int32{1500, 1500},
			out:         -2,
			expError:    ErrInvalidOutcome,
			expRatingsA: []int32{1500, 1500},
			expRatingsB: []int32{1500, 1500},
		},
		{
			description: "invalid outcome. Too big.",
			ratingsA:    []int32{1500, 1500},
			ratingsB:    []int32{1500, 1500},
			out:         3,
			expError:    ErrInvalidOutcome,
			expRatingsA: []int32{1500, 1500},
			expRatingsB: []int32{1500, 1500},
		},
		{
			description: "Same ratings. Team A wins.",
			ratingsA:    []int32{1500, 1500},
			ratingsB:    []int32{1500, 1500},
			out:         0,
			expError:    nil,
			expRatingsA: []int32{1508, 1508},
			expRatingsB: []int32{1492, 1492},
		},
		{
			description: "Same ratings. Team B wins.",
			ratingsA:    []int32{1500, 1500},
			ratingsB:    []int32{1500, 1500},
			out:         1,
			expError:    nil,
			expRatingsA: []int32{1492, 1492},
			expRatingsB: []int32{1508, 1508},
		},
		{
			description: "Same ratings. Draw.",
			ratingsA:    []int32{1500, 1500},
			ratingsB:    []int32{1500, 1500},
			out:         2,
			expError:    nil,
			expRatingsA: []int32{1500, 1500},
			expRatingsB: []int32{1500, 1500},
		},
		{
			description: "Different ratings. Team A wins.",
			ratingsA:    []int32{1500, 1800},
			ratingsB:    []int32{1400, 1600},
			out:         0,
			expError:    nil,
			expRatingsA: []int32{1505, 1804},
			expRatingsB: []int32{1396, 1595},
		},
		{
			description: "Different ratings. Team B wins.",
			ratingsA:    []int32{1500, 1800},
			ratingsB:    []int32{1400, 1600},
			out:         1,
			expError:    nil,
			expRatingsA: []int32{1490, 1788},
			expRatingsB: []int32{1412, 1610},
		},
		{
			description: "Different ratings. Draw.",
			ratingsA:    []int32{1500, 1800},
			ratingsB:    []int32{1400, 1600},
			out:         2,
			expError:    nil,
			expRatingsA: []int32{1497, 1797},
			expRatingsB: []int32{1403, 1603},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			elo := NewEloDefault()
			newRA, newRB, err := elo.GetNewRatingsTeams(tc.ratingsA, tc.ratingsB, tc.out)
			assert.ErrorIs(t, err, tc.expError)
			assert.Equal(t, tc.expRatingsA, newRA)
			assert.Equal(t, tc.expRatingsB, newRB)
		})
	}
}

type sCalculatorMock struct{}

func (c *sCalculatorMock) getSValue(_ int32, _ Outcome) float64 {
	return 1
}
