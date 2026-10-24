package history

import (
	"math"
	"sort"
)

const (
	// SessionGapHours is the git-hours session gap.
	SessionGapHours = 4.0
	// FirstCommitHours is the credit for the first commit of each session.
	FirstCommitHours = 0.5
)

// ActiveHours is author active hours (git-hours sessions) over author
// timestamps in Unix seconds: gaps <= 4 h add the gap, a longer gap starts a
// session, 0.5 h credit per session. The summation order is fixed so results
// are reproducible bit for bit. Empty input yields the 0.5 h floor.
func ActiveHours(authorUnix []float64) float64 {
	t := append([]float64(nil), authorUnix...)
	sort.Float64s(t)
	h := FirstCommitHours
	for i := 0; i+1 < len(t); i++ {
		gap := (t[i+1] - t[i]) / 3600.0
		if gap <= SessionGapHours {
			h += gap
		} else {
			h += FirstCommitHours
		}
	}
	return h
}

// percentile is the linear-interpolation percentile (numpy default); q in [0,100].
// xs must be non-empty.
func percentile(xs []float64, q float64) float64 {
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	if len(s) == 1 {
		return s[0]
	}
	pos := q / 100 * float64(len(s)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return s[lo] + (s[hi]-s[lo])*(pos-float64(lo))
}
