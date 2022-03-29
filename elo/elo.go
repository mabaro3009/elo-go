package elo

import (
	"errors"
	"math"
)

type SCalculator interface {
	getSValue(n int32, o Outcome) float64
}

type KCalculator interface {
	getKFactor(rating int32) float64
}

const (
	DefaultKFactor = 32
	DefaultDValue  = 400

	Win Outcome = iota
	Loss
	Draw
)

var (
	ErrInvalidOutcome = errors.New("invalid Outcome, must be 0, 1 or 2")
)

type Outcome int

type Elo struct {
	dValue float64
	s      SCalculator
	k      KCalculator
}

// NewElo returns a new elo object.
// kFactor is used to determine how much a player's rating can change after a match.
// dValue affects how the difference in ratings translates to win probabilities.
func NewElo(dValue float64, s SCalculator, k KCalculator) *Elo {
	return &Elo{
		dValue: dValue,
		s:      s,
		k:      k,
	}
}

// NewEloDefault returns a new Elo object with default fields.
func NewEloDefault() *Elo {
	return &Elo{
		dValue: DefaultDValue,
		s:      NewSCalculatorLinear(),
		k:      NewKCalculatorConst(DefaultKFactor),
	}
}

// GetExpectedScore returns the expected Outcome of the game.
// ratingA is the elo rating of playerA and ratingB is the elo rating of playerB
// precision determines with how many decimals the expected scores are returned.
// Return value ranges from 0 to 1. A value of 0.75 indicates that playerA has an expected 75% chance of winning.
func (e *Elo) GetExpectedScore(ratingA, ratingB, precision int32) float64 {
	return toFixed(1/(1+math.Pow(10, float64(ratingB-ratingA)/e.dValue)), precision)
}

// GetNewRatings returns the new rating for playerA
// ratingA is the elo rating of playerA and ratingB is the elo rating of playerB
// Outcome is the result of the match. 0 for playerA Win, 1 for playerB Win and 2 for Draw.
func (e *Elo) GetNewRatings(ratingA, ratingB, out int32) (int32, int32, error) {
	if out < 0 || out > 2 {
		return ratingA, ratingB, ErrInvalidOutcome
	}

	return e.getNewRating(ratingA, ratingB, e.s.getSValue(2, getOutcome(0, out))), e.getNewRating(ratingB, ratingA, e.s.getSValue(2, getOutcome(1, out))), nil
}

func getOutcome(p, o int32) Outcome {
	switch o {
	case 0:
		if p == 0 {
			return Win
		}
		return Loss

	case 1:
		if p == 0 {
			return Loss
		}
		return Win

	default:
		return Draw
	}
}

func (e *Elo) getNewRating(ratingA, ratingB int32, s float64) int32 {
	expScore := e.GetExpectedScore(ratingA, ratingB, 0)

	return ratingA + int32(e.k.getKFactor(ratingA)*(s-expScore))
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int32) float64 {
	if precision == 0 {
		return num
	}
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
