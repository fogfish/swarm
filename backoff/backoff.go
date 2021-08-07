package backoff

import (
	"math/rand"
	"time"
)

/*

Seq is a sequence of delays
*/
type Seq []time.Duration

/*

Const is a sequence of constant delays
*/
func Const(delay time.Duration, n int) Seq {
	seq := make([]time.Duration, n)
	seq[0] = delay

	for i := 1; i < n; i++ {
		seq[i] = delay
	}

	return seq
}

/*

Linear is a sequence of constant delays
*/
func Linear(delay time.Duration, n int) Seq {
	seq := make([]time.Duration, n)
	seq[0] = delay

	for i := 1; i < n; i++ {
		seq[i] = seq[i-1] + delay
	}

	return seq
}

/*

Exp is a sequence of exponential delays
*/
func Exp(delay time.Duration, n int, factor float64) Seq {
	seq := make([]time.Duration, n)
	seq[0] = delay
	for i := 1; i < n; i++ {
		seq[i] = seq[i-1] + interval(factor, rand.Float64(), seq[i-1])
	}
	return seq
}

func interval(randomizationFactor, random float64, currentInterval time.Duration) time.Duration {
	var delta = randomizationFactor * float64(currentInterval)
	var minInterval = float64(currentInterval) - delta
	var maxInterval = float64(currentInterval) + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	return time.Duration(minInterval + (random * (maxInterval - minInterval + 1)))
}

/*

Deadline defines a total time for the delay
*/
func (seq Seq) Deadline(t time.Duration) Seq {
	var sum time.Duration

	for i := 0; i < len(seq); i++ {
		sum = sum + seq[i]
		if sum > t {
			return seq[0:i]
		}
	}

	return seq
}

/*

Retry function
*/
func (seq Seq) Retry(f func() error) (err error) {
	for _, t := range seq {
		if err = f(); err == nil {
			return
		}
		time.Sleep(t)
	}
	return
}
