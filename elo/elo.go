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

// GetNewRatings returns the new rating for playerA and playerB
// ratingA is the elo rating of playerA and ratingB is the elo rating of playerB
// Outcome is the result of the match. 0 for playerA Win, 1 for playerB Win and 2 for Draw.
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
// ratingsA are the elo ratings of players in teamA and ratingsB are the elo rating of players in teamB.
// Outcome is the result of the match. 0 for teamA Win, 1 for teamB Win and 2 for Draw.
func (e *Elo) GetNewRatingsTeams(ratingsA, ratingsB []int, out int) ([]int, []int, error) {
	if out < 0 || out > 2 {
		return ratingsA, ratingsB, ErrInvalidOutcome
	}

	avgA := getAverage(ratingsA)
	avgB := getAverage(ratingsB)

	outcomeA := getOutcome(0, out, 2)
	outcomeB := getOutcome(1, out, 2)

	modifierA := getModifier(len(ratingsA), len(ratingsB))
	modifierB := getModifier(len(ratingsB), len(ratingsA))

	incrementA := e.getIncrement(avgA, avgB, 2, e.s.getSValue(2, outcomeA), modifierA, modifierB)
	newRatingsA := e.getNewIndividualRatings(incrementA, ratingsA)

	incrementB := e.getIncrement(avgB, avgA, 2, e.s.getSValue(2, outcomeB), modifierB, modifierA)
	newRatingsB := e.getNewIndividualRatings(incrementB, ratingsB)

	return newRatingsA, newRatingsB, nil
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

func getAverageExcluding(in []int, i int) int {
	excl := make([]int, len(in))
	copy(excl, in)
	excl[i] = excl[len(excl)-1]

	return getAverage(excl[:len(excl)-1])
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
