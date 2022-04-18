package elo

import (
	"errors"
	"math"
	"sort"
)

type SCalculator interface {
	getSValue(n int, o Outcome) float64
}

type KCalculator interface {
	getKFactor(rating int) float64
}

const (
	DefaultKFactor = 32
	DefaultDValue  = 400
)

const (
	Win Outcome = iota
	Loss
	Draw
)

var (
	ErrInvalidOutcome = errors.New("invalid outcome, must be 0, 1 or 2")
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
func (e *Elo) GetExpectedScore(ratingA, ratingB, precision int) float64 {
	return toFixed(1/(1+math.Pow(10, float64(ratingB-ratingA)/e.dValue)), precision)
}

// GetNewRatings returns the new rating for playerA and playerB.
// Each element of ratings represents the rating of each player. ratings[i] is the rating of player i.
// Outcome is the result of the match. i for playerI Win and len(ratings) for Draw.
func (e *Elo) GetNewRatings(ratings []int, out int) ([]int, error) {
	if out < 0 || out > len(ratings) {
		return ratings, ErrInvalidOutcome
	}

	n := len(ratings)
	newRatings := make([]int, 0, n)
	for i, rating := range ratings {
		avg := getAverageExcluding(ratings, i)
		outcome := getOutcome(i, out, n)
		newRatings = append(newRatings, e.getNewRating(rating, avg, len(ratings), e.s.getSValue(n, outcome)))
	}

	return newRatings, nil
}

// GetNewRatingsTeams returns the new ratings for each player in a team match.
// Each slice in matrix ratings represents the ratings of each team. ratings[i] are the ratings of team i.
// Outcome is the result of the match. i for teamI win, len(ratings) for draw.
func (e *Elo) GetNewRatingsTeams(ratings [][]int, out int) ([][]int, error) {
	if out < 0 || out > len(ratings) {
		return ratings, ErrInvalidOutcome
	}

	n := len(ratings)
	newRatings := make([][]int, 0, n)
	for i, ratingsTeam := range ratings {
		avg := getAverage(ratingsTeam)
		avgRest := getMatrixAverageExcluding(ratings, i)
		outcome := getOutcome(i, out, n)

		avgLen := getAverageLenExcluding(ratings, i)
		modifier := getModifier(len(ratingsTeam), avgLen)
		modifierRest := getModifier(avgLen, len(ratingsTeam))
		increment := e.getIncrement(avg, avgRest, n, e.s.getSValue(2, outcome), modifier, modifierRest)

		newRatings = append(newRatings, e.getNewIndividualRatings(increment, ratingsTeam))
	}

	return newRatings, nil
}

func getOutcome(i, o, l int) Outcome {
	if i == o {
		return Win
	}
	if o == l {
		return Draw
	}

	return Loss
}

func (e *Elo) getNewRating(ratingA, ratingB, n int, s float64) int {
	return ratingA + e.getIncrement(ratingA, ratingB, n, s, 1, 1)
}

type sortedRating struct {
	originalIndex int
	rating        int
}

type sortedRatings []sortedRating

func (s sortedRatings) Len() int           { return len(s) }
func (s sortedRatings) Less(i, j int) bool { return s[i].rating < s[j].rating }
func (s sortedRatings) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (e *Elo) getNewIndividualRatings(totalIncrement int, ratings []int) []int {
	orderedRatingsWithIndex := make(sortedRatings, 0, len(ratings))
	for i, r := range ratings {
		orderedRatingsWithIndex = append(orderedRatingsWithIndex, sortedRating{
			originalIndex: i,
			rating:        r,
		})
	}
	if totalIncrement >= 0 {
		sort.Sort(orderedRatingsWithIndex)
	} else {
		sort.Sort(sort.Reverse(orderedRatingsWithIndex))
	}

	orderedRatings := make([]int, len(orderedRatingsWithIndex))
	for i, r := range orderedRatingsWithIndex {
		orderedRatings[i] = r.rating
	}

	incrementRest := totalIncrement
	for i := range orderedRatingsWithIndex {
		ratio := getRatio(i, orderedRatings, totalIncrement >= 0)
		var increment int
		if totalIncrement >= 0 {
			increment = int(math.Floor(ratio * float64(totalIncrement)))
		} else {
			increment = int(math.Ceil(ratio * float64(totalIncrement)))
		}
		orderedRatingsWithIndex[i].rating += increment
		incrementRest -= increment
	}

	for i := 0; incrementRest != 0; i++ {
		index := i % len(orderedRatingsWithIndex)
		if totalIncrement >= 0 {
			orderedRatingsWithIndex[index].rating++
			incrementRest--
		} else {
			orderedRatingsWithIndex[index].rating--
			incrementRest++
		}
	}

	newRatings := make([]int, len(ratings))
	for _, rating := range orderedRatingsWithIndex {
		newRatings[rating.originalIndex] = rating.rating
	}

	return newRatings
}

func (e *Elo) getIncrement(ratingA, ratingB, n int, s, mA, mB float64) int {
	expScore := e.GetExpectedScore(int(float64(ratingA)*mA), int(float64(ratingB)*mB), 0)

	return int(e.k.getKFactor(ratingA)*(s-expScore)) * (n - 1)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	if precision == 0 {
		return num
	}
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func getMatrixAverageExcluding(in [][]int, i int) int {
	averages := make([]int, 0, len(in))
	for j, aux := range in {
		if i == j {
			continue
		}
		averages = append(averages, getAverage(aux))
	}

	return getAverage(averages)
}

func getAverageExcluding(in []int, i int) int {
	excl := make([]int, len(in))
	copy(excl, in)
	excl[i] = excl[len(excl)-1]

	return getAverage(excl[:len(excl)-1])
}

func getAverageLenExcluding(in [][]int, i int) int {
	excl := make([]int, 0, len(in))
	for j, v := range in {
		if i == j {
			continue
		}
		excl = append(excl, len(v))
	}

	return getAverage(excl)
}

func getAverage(in []int) int {
	return getSum(in) / len(in)
}

func getRatio(i int, sl []int, gained bool) float64 {
	var index int
	if gained {
		index = len(sl) - 1 - i
	} else {
		index = i
	}
	sum := getSum(sl)
	a := sl[index]
	return float64(a) / float64(sum)
}

func getSum(in []int) int {
	var sum int
	for _, a := range in {
		sum += a
	}

	return sum
}

func getModifier(lenA, lenB int) float64 {
	return float64(lenA) / float64(lenB)
}
