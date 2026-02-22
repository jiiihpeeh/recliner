package hooks

import (
	"math"
	"time"
)

type EasingFunc func(float64) float64

var (
	EasingLinear    EasingFunc = func(t float64) float64 { return t }
	EasingInQuad    EasingFunc = func(t float64) float64 { return t * t }
	EasingOutQuad   EasingFunc = func(t float64) float64 { return t * (2 - t) }
	EasingInOutQuad EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	}
	EasingInCubic  EasingFunc = func(t float64) float64 { return t * t * t }
	EasingOutCubic EasingFunc = func(t float64) float64 {
		t--
		return t*t*t + 1
	}
	EasingBounce EasingFunc = func(t float64) float64 {
		if t < (1 / 2.75) {
			return 7.5625 * t * t
		} else if t < (2 / 2.75) {
			t -= (1.5 / 2.75)
			return 7.5625*t*t + 0.75
		} else if t < (2.5 / 2.75) {
			t -= (2.25 / 2.75)
			return 7.5625*t*t + 0.9375
		} else {
			t -= (2.625 / 2.75)
			return 7.5625*t*t + 0.984375
		}
	}
)

type AnimationOptions struct {
	Duration time.Duration
	Easing   EasingFunc
	Delay    time.Duration
}

// UseAnimation triggers a value transition from 0 to 1 over a duration.
// It returns the current animated value (float64) and a function to reset/start the animation.
func UseAnimation(hc *HooksContext, active bool, opts AnimationOptions) float64 {
	if opts.Easing == nil {
		opts.Easing = EasingLinear
	}
	if opts.Duration == 0 {
		opts.Duration = 300 * time.Millisecond
	}

	startVal := 0.0
	if !active {
		startVal = 0.0
	}

	val, setVal := UseState(hc, startVal)
	startTime, setStartTime := UseState(hc, time.Time{})

	UseEffect(hc, func() func() {
		if !active {
			setVal(0.0)
			setStartTime(time.Time{})
			return nil
		}

		if startTime.IsZero() {
			setStartTime(time.Now().Add(opts.Delay))
		}

		var ticker *time.Ticker
		// We use a relatively high frequency ticker for smooth terminal animations (30-60fps)
		ticker = time.NewTicker(33 * time.Millisecond)
		stop := make(chan struct{})

		go func() {
			for {
				select {
				case now := <-ticker.C:
					if startTime.IsZero() || now.Before(startTime) {
						continue
					}

					elapsed := now.Sub(startTime)
					progress := float64(elapsed) / float64(opts.Duration)

					if progress >= 1.0 {
						setVal(1.0)
						ticker.Stop()
						return
					}

					setVal(opts.Easing(progress))
				case <-stop:
					return
				}
			}
		}()

		return func() {
			ticker.Stop()
			close(stop)
		}
	}, []any{active, opts.Duration})

	return val
}

// Lerp provides a simple linear interpolation for animated values
func Lerp(start, end, t float64) float64 {
	return start + (end-start)*t
}

// LerpInt provides a simple linear interpolation for integer values (useful for layout)
func LerpInt(start, end int, t float64) int {
	return int(math.Round(float64(start) + float64(end-start)*t))
}
