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
	ErrInvalidOutcome   = errors.New("invalid Outcome, must be 0, 1 or 2")
	ErrTeamLenMissmatch = errors.New("teams are not of the same length")
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

// GetNewRatings returns the new rating for playerA and playerB
// ratingA is the elo rating of playerA and ratingB is the elo rating of playerB
// Outcome is the result of the match. 0 for playerA Win, 1 for playerB Win and 2 for Draw.
func (e *Elo) GetNewRatings(ratingA, ratingB, out int32) (int32, int32, error) {
	if out < 0 || out > 2 {
		return ratingA, ratingB, ErrInvalidOutcome
	}

	return e.getNewRating(ratingA, ratingB, e.s.getSValue(2, getOutcome(0, out))), e.getNewRating(ratingB, ratingA, e.s.getSValue(2, getOutcome(1, out))), nil
}

// GetNewRatingsTeams returns the new ratings for each player in a team match.
// ratingsA are the elo ratings of players in teamA and ratingsB are the elo rating of players in teamB.
// Outcome is the result of the match. 0 for teamA Win, 1 for teamB Win and 2 for Draw.
func (e *Elo) GetNewRatingsTeams(ratingsA, ratingsB []int32, out int32) ([]int32, []int32, error) {
	if len(ratingsA) != len(ratingsB) {
		return ratingsA, ratingsB, ErrTeamLenMissmatch
	}
	if out < 0 || out > 2 {
		return ratingsA, ratingsB, ErrInvalidOutcome
	}

	avgA := getAverage(ratingsA)
	avgB := getAverage(ratingsB)

	outcomeA := getOutcome(0, out)
	outcomeB := getOutcome(1, out)

	incrementA := e.getIncrement(avgA, avgB, e.s.getSValue(2, outcomeA))
	newRatingsA := make([]int32, 0, len(ratingsA))
	for i := range ratingsA {
		newRatingsA = append(newRatingsA, e.getNewIndividualRating(incrementA, i, ratingsA, outcomeA))
	}

	incrementB := e.getIncrement(avgB, avgA, e.s.getSValue(2, outcomeB))
	newRatingsB := make([]int32, 0, len(ratingsB))
	for i := range ratingsB {
		newRatingsB = append(newRatingsB, e.getNewIndividualRating(incrementB, i, ratingsB, outcomeB))
	}

	return newRatingsA, newRatingsB, nil
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
	return ratingA + e.getIncrement(ratingA, ratingB, s)
}

func (e *Elo) getNewIndividualRating(totalIncrement int32, i int, ratings []int32, o Outcome) int32 {
	ratio := getRatio(i, ratings, o)
	increment := math.Round(ratio * float64(totalIncrement))

	return ratings[i] + int32(increment)
}

func (e *Elo) getIncrement(ratingA, ratingB int32, s float64) int32 {
	expScore := e.GetExpectedScore(ratingA, ratingB, 0)

	return int32(e.k.getKFactor(ratingA) * (s - expScore))
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

func getAverage(in []int32) int32 {
	return getSum(in) / int32(len(in))
}

func getRatio(i int, sl []int32, o Outcome) float64 {
	var index int
	switch o {
	case Loss:
		index = i
	default:
		index = len(sl) - 1 - i
	}
	sum := getSum(sl)
	a := sl[index]
	return float64(a) / float64(sum)
}

func getSum(in []int32) int32 {
	var sum int32
	for _, a := range in {
		sum += a
	}

	return sum
}
