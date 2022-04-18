package elo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetExpectedScore(t *testing.T) {
	testcases := []struct {
		description string
		d           float64
		p           int
		ra          int
		rb          int
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
		r           []int
		outcome     int
		s           SCalculator
		expErr      error
		expNewR     []int
	}{
		{
			description: "negative Outcome",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{1500, 1500},
			outcome:     -1,
			s:           NewSCalculatorLinear(),
			expErr:      ErrInvalidOutcome,
			expNewR:     []int{1500, 1500},
		},
		{
			description: "invalid Outcome",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{1500, 1500},
			outcome:     3,
			s:           NewSCalculatorLinear(),
			expErr:      ErrInvalidOutcome,
			expNewR:     []int{1500, 1500},
		},
		{
			description: "same rating, playerA wins",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{1500, 1500},
			outcome:     0,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewR:     []int{1516, 1484},
		},
		{
			description: "same rating, playerA wins with non-default k factor",
			k:           NewKCalculatorConst(40),
			r:           []int{1500, 1500},
			outcome:     0,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewR:     []int{1520, 1480},
		},
		{
			description: "same rating, playerB wins",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{1500, 1500},
			outcome:     1,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewR:     []int{1484, 1516},
		},
		{
			description: "same rating, Draw",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{1500, 1500},
			outcome:     2,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewR:     []int{1500, 1500},
		},
		{
			description: "same rating, with mock s",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{1500, 1500},
			outcome:     2,
			s:           &sCalculatorMock{},
			expErr:      nil,
			expNewR:     []int{1516, 1516},
		},
		{
			description: "ratingA is higher, playerA wins",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{2200, 1900},
			outcome:     0,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewR:     []int{2204, 1896},
		},
		{
			description: "ratingA is higher, playerB wins",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{2200, 1900},
			outcome:     1,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewR:     []int{2173, 1927},
		},
		{
			description: "ratingA is higher, Draw",
			k:           NewKCalculatorConst(DefaultKFactor),
			r:           []int{2200, 1900},
			outcome:     2,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewR:     []int{2189, 1911},
		},
		{
			description: "ratingA is higher, a wins, USCF kFactor",
			k:           NewKCalculatorUSCF(),
			r:           []int{2200, 1900},
			outcome:     2,
			s:           NewSCalculatorLinear(),
			expErr:      nil,
			expNewR:     []int{2192, 1911},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			elo := NewElo(DefaultDValue, tc.s, tc.k)

			newRatings, err := elo.GetNewRatings(tc.r, tc.outcome)
			assert.ErrorIs(t, err, tc.expErr)
			assert.Equal(t, tc.expNewR, newRatings)
		})
	}
}

func TestGetNewRatingsTeams(t *testing.T) {
	testCases := []struct {
		description string
		ratingsA    []int
		ratingsB    []int
		out         int
		expError    error
		expRatingsA []int
		expRatingsB []int
	}{
		{
			description: "invalid outcome. Negative.",
			ratingsA:    []int{1500, 1500},
			ratingsB:    []int{1500, 1500},
			out:         -2,
			expError:    ErrInvalidOutcome,
			expRatingsA: []int{1500, 1500},
			expRatingsB: []int{1500, 1500},
		},
		{
			description: "invalid outcome. Too big.",
			ratingsA:    []int{1500, 1500},
			ratingsB:    []int{1500, 1500},
			out:         3,
			expError:    ErrInvalidOutcome,
			expRatingsA: []int{1500, 1500},
			expRatingsB: []int{1500, 1500},
		},
		{
			description: "Same ratings. Team A wins.",
			ratingsA:    []int{1500, 1500},
			ratingsB:    []int{1500, 1500},
			out:         0,
			expError:    nil,
			expRatingsA: []int{1508, 1508},
			expRatingsB: []int{1492, 1492},
		},
		{
			description: "Same ratings. Team B wins.",
			ratingsA:    []int{1500, 1500},
			ratingsB:    []int{1500, 1500},
			out:         1,
			expError:    nil,
			expRatingsA: []int{1492, 1492},
			expRatingsB: []int{1508, 1508},
		},
		{
			description: "Same ratings. Draw.",
			ratingsA:    []int{1500, 1500},
			ratingsB:    []int{1500, 1500},
			out:         2,
			expError:    nil,
			expRatingsA: []int{1500, 1500},
			expRatingsB: []int{1500, 1500},
		},
		{
			description: "Different ratings. Team A wins.",
			ratingsA:    []int{1500, 1800},
			ratingsB:    []int{1400, 1600},
			out:         0,
			expError:    nil,
			expRatingsA: []int{1505, 1804},
			expRatingsB: []int{1396, 1595},
		},
		{
			description: "Different ratings. Team B wins.",
			ratingsA:    []int{1500, 1800},
			ratingsB:    []int{1400, 1600},
			out:         1,
			expError:    nil,
			expRatingsA: []int{1490, 1788},
			expRatingsB: []int{1412, 1610},
		},
		{
			description: "Different ratings. Draw.",
			ratingsA:    []int{1500, 1800},
			ratingsB:    []int{1400, 1600},
			out:         2,
			expError:    nil,
			expRatingsA: []int{1498, 1796},
			expRatingsB: []int{1404, 1602},
		},
		{
			description: "team length mismatch. Team A wins",
			ratingsA:    []int{1500, 1500, 1500},
			ratingsB:    []int{1500, 1500},
			out:         0,
			expError:    nil,
			expRatingsA: []int{1500, 1500, 1500},
			expRatingsB: []int{1500, 1500},
		},
		{
			description: "team length mismatch. Team B wins",
			ratingsA:    []int{1500, 1500, 1500},
			ratingsB:    []int{1500, 1500},
			out:         1,
			expError:    nil,
			expRatingsA: []int{1489, 1490, 1490},
			expRatingsB: []int{1516, 1515},
		},
		{
			description: "team length mismatch. Draw",
			ratingsA:    []int{1500, 1500, 1500},
			ratingsB:    []int{1500, 1500},
			out:         2,
			expError:    nil,
			expRatingsA: []int{1495, 1495, 1495},
			expRatingsB: []int{1508, 1507},
		},
		{
			description: "team length mismatch and different ratings. Team A wins",
			ratingsA:    []int{800, 900, 960},
			ratingsB:    []int{2000, 2100},
			out:         0,
			expError:    nil,
			expRatingsA: []int{807, 905, 965},
			expRatingsB: []int{1992, 2091},
		},
		{
			description: "team length mismatch and different ratings. Team B wins",
			ratingsA:    []int{900, 800, 960},
			ratingsB:    []int{2000, 2100},
			out:         1,
			expError:    nil,
			expRatingsA: []int{896, 796, 954},
			expRatingsB: []int{2008, 2106},
		},
		{
			description: "team length mismatch and different ratings. Draw",
			ratingsA:    []int{900, 800, 960},
			ratingsB:    []int{2000, 2100},
			out:         2,
			expError:    nil,
			expRatingsA: []int{900, 801, 960},
			expRatingsB: []int{2000, 2099},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			elo := NewEloDefault()

			newRatings, err := elo.GetNewRatingsTeams([][]int{tc.ratingsA, tc.ratingsB}, tc.out)
			assert.ErrorIs(t, err, tc.expError)
			assert.Equal(t, tc.expRatingsA, newRatings[0])
			assert.Equal(t, tc.expRatingsB, newRatings[1])
		})
	}
}

type sCalculatorMock struct{}

func (c *sCalculatorMock) getSValue(_ int, _ Outcome) float64 {
	return 1
}
