# Elo rating in Go
This library implements an Elo rating system in Golang.

## Usage
The Elo system includes three factors that can be modified: *kfactor*, *d* and *s*.

The *d* value is always a constant and the *kfactor* and *s* values can be calculated via different implementations.
This library includes some common implementations for both, but the user may code their own.

Example usage with default values:
```
// This initializes an elo object with:
// constant k-factor of 32
// d = 400
// linear implementation of s (1 for win, 0 for loss, 1/n for draw)
e := elo.NewEloDefault()

// Get the probability of a player with a rating of 1750 to beat 
// a player with a rating of 1500 with 2 decimals of precission.
expScore := e.GetExpectedScore(1750, 1500, 2)

// Get the new ratings when a player of rating of 1750 beats
// a player with a rating of 1500.
// Use 0 to indicate the first player wins, 1 for the second player
// and 2 for a draw.
newRatingA, newRatingB, err := e.GetNewRatings(1750, 1500, 0)


teamA := []int32{1000, 1500}
teamB := []int32{1300, 1400}

// Get the new ratings when teamA beats teamB in a 2vs2 match.
newRatingsA, newRatingsB, err := e.GetNewRatingsTeams(teamA, teamB, 0)
```

Example initializacion with custom values:
```

// kCalculatorCustom is an object that implements a custom k function
type kCalculatorCustom struct {
    minValue float64
    maxValue float64
    threshhold int32
}

func newKCalculatorCustom(minValue, maxValue float64, threshhold int32) *kCalculatorCustom {
    return &kCalculatorCustom {
        minValue: minValue,
        maxValue: maxValue,
        threshhold: threshhold,
    }
}

func (c *kCalculatorCustom) getKFactor(rating int32) float64 {
    if rating < c.threshhold {
        return c.minValue
    }
    
    return c.maxValue
}

e := nelo.NewElo(200, elo.NewSCalculatorLinear(), newKCalculatorCustom(20, 40, 1500))
```
