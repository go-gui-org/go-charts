package axis

import (
	"cmp"
	"math"
	"time"

	"github.com/go-gui-org/go-charts/internal/fmath"
)

// Time is a time-based axis.
type Time struct {
	title          string
	min, max       time.Time
	format         string
	autoFormat     bool
	tickFormat     TickFormat
	overrideDomain bool
}

// TimeCfg configures a time axis.
type TimeCfg struct {
	Title  string
	Min    time.Time
	Max    time.Time
	Format string // Go time format string

	// TickFormat overrides the default time formatting. The
	// function receives seconds (float64) as produced by
	// timeToSeconds.
	TickFormat TickFormat
}

// NewTime creates a time axis.
func NewTime(cfg TimeCfg) *Time {
	return &Time{
		title:      cfg.Title,
		min:        cfg.Min,
		max:        cfg.Max,
		format:     cmp.Or(cfg.Format, "2006-01-02"),
		autoFormat: cfg.Format == "",
		tickFormat: cfg.TickFormat,
	}
}

// FormatTime returns a TickFormat that converts seconds (float64)
// to a string using the given Go time layout.
func FormatTime(layout string) TickFormat {
	return func(seconds float64) string {
		sec := int64(seconds)
		nsec := int64((seconds - float64(sec)) * 1e9)
		return time.Unix(sec, nsec).Format(layout)
	}
}

// Label implements Axis.
func (a *Time) Label() string { return a.title }

// SetRange implements Axis. Values are seconds (float64).
// Non-finite values are ignored.
func (a *Time) SetRange(min, max float64) {
	if !fmath.Finite(min) || !fmath.Finite(max) {
		return
	}
	a.min = secondsToTime(min)
	a.max = secondsToTime(max)
}

// Domain implements Axis. Returns seconds (float64).
func (a *Time) Domain() (float64, float64) {
	return timeToSeconds(a.min), timeToSeconds(a.max)
}

// SetOverrideDomain implements Axis.
func (a *Time) SetOverrideDomain(v bool) { a.overrideDomain = v }

// timeStep defines a candidate tick step for time axes.
type timeStep struct {
	seconds    float64 // 0 for month/year steps
	months     int     // non-zero for month/year steps
	autoFormat string
}

// timeSteps lists candidate steps from smallest to largest.
// Month/year steps use the months field; fixed-duration steps
// use seconds.
var timeSteps = []timeStep{
	{seconds: 5, autoFormat: "15:04:05"},
	{seconds: 10, autoFormat: "15:04:05"},
	{seconds: 15, autoFormat: "15:04:05"},
	{seconds: 30, autoFormat: "15:04:05"},
	{seconds: 60, autoFormat: "15:04"},
	{seconds: 5 * 60, autoFormat: "15:04"},
	{seconds: 10 * 60, autoFormat: "15:04"},
	{seconds: 15 * 60, autoFormat: "15:04"},
	{seconds: 30 * 60, autoFormat: "15:04"},
	{seconds: 3600, autoFormat: "15:04"},
	{seconds: 2 * 3600, autoFormat: "15:04"},
	{seconds: 3 * 3600, autoFormat: "15:04"},
	{seconds: 6 * 3600, autoFormat: "Jan 02 15:04"},
	{seconds: 12 * 3600, autoFormat: "Jan 02 15:04"},
	{seconds: 86400, autoFormat: "Jan 02"},
	{seconds: 2 * 86400, autoFormat: "Jan 02"},
	{seconds: 7 * 86400, autoFormat: "Jan 02"},
	{seconds: 14 * 86400, autoFormat: "Jan 02"},
	{months: 1, autoFormat: "Jan 2006"},
	{months: 2, autoFormat: "Jan 2006"},
	{months: 3, autoFormat: "Jan 2006"},
	{months: 6, autoFormat: "Jan 2006"},
	{months: 12, autoFormat: "2006"},
	{months: 24, autoFormat: "2006"},
	{months: 60, autoFormat: "2006"},
	{months: 120, autoFormat: "2006"},
}

const targetTicks = 8

// Ticks implements Axis. Generates time-aligned ticks with
// adaptive formatting based on the time range.
func (a *Time) Ticks(pixelMin, pixelMax float32) []Tick {
	minSec := timeToSeconds(a.min)
	maxSec := timeToSeconds(a.max)
	if !fmath.Finite(minSec) || !fmath.Finite(maxSec) {
		return nil
	}
	dur := maxSec - minSec
	if dur <= 0 {
		return nil
	}

	// Pick the step yielding closest to targetTicks.
	bestIdx := 0
	bestDiff := float64(1e18)
	for i, ts := range timeSteps {
		var n float64
		if ts.months > 0 {
			n = dur / (float64(ts.months) * 30.44 * 86400)
		} else {
			n = dur / ts.seconds
		}
		diff := math.Abs(n - targetTicks)
		if diff < bestDiff {
			bestDiff = diff
			bestIdx = i
		}
	}
	step := timeSteps[bestIdx]

	// Determine label format.
	labelFmt := a.format
	if a.autoFormat {
		labelFmt = step.autoFormat
	}

	// Align first tick to a natural boundary.
	var first time.Time
	if step.months > 0 {
		first = alignMonth(a.min, step.months)
	} else {
		first = alignFixed(a.min, step.seconds)
	}

	const maxTickCount = 500
	ticks := make([]Tick, 0, min(targetTicks*2, maxTickCount))

	cur := first
	for range maxTickCount {
		sec := timeToSeconds(cur)
		if sec > maxSec+1e-9 {
			break
		}
		if sec >= minSec-1e-9 {
			label := cur.Format(labelFmt)
			if a.tickFormat != nil {
				label = a.tickFormat(sec)
			}
			ticks = append(ticks, Tick{
				Value:    sec,
				Label:    label,
				Position: a.Transform(sec, pixelMin, pixelMax),
			})
		}
		if step.months > 0 {
			cur = cur.AddDate(0, step.months, 0)
		} else {
			cur = cur.Add(
				time.Duration(step.seconds * float64(time.Second)))
		}
	}
	return ticks
}

// alignMonth snaps t to the start of a month boundary aligned
// to the given month step.
func alignMonth(t time.Time, months int) time.Time {
	if months <= 0 {
		return t
	}
	y, m, _ := t.Date()
	loc := t.Location()
	mi := int(m) - 1 // 0-based month index
	mi = (mi / months) * months
	return time.Date(y, time.Month(mi+1), 1, 0, 0, 0, 0, loc)
}

// alignFixed snaps t to the nearest step boundary at or before t.
func alignFixed(t time.Time, stepSec float64) time.Time {
	if stepSec <= 0 {
		return t
	}
	loc := t.Location()
	// Align relative to midnight of the day for sub-day steps,
	// working in the original timezone.
	ref := time.Date(
		t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0, loc)
	elapsed := t.Sub(ref).Seconds()
	aligned := float64(int64(elapsed/stepSec)) * stepSec
	return ref.Add(time.Duration(aligned * float64(time.Second)))
}

// timeToSeconds converts a time.Time to seconds as float64,
// avoiding int64 overflow from UnixNano.
func timeToSeconds(t time.Time) float64 {
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}

// secondsToTime converts seconds (float64) to time.Time.
// Non-finite inputs return the zero time.
func secondsToTime(sec float64) time.Time {
	if !fmath.Finite(sec) {
		return time.Time{}
	}
	s := int64(sec)
	ns := int64((sec - float64(s)) * 1e9)
	return time.Unix(s, ns)
}

// Transform implements Axis. Value is expected as seconds
// (float64), matching the output of timeToSeconds.
func (a *Time) Transform(value float64, pixelMin, pixelMax float32) float32 {
	minSec := timeToSeconds(a.min)
	rangeSec := timeToSeconds(a.max) - minSec
	if rangeSec == 0 {
		return pixelMin
	}
	t := (value - minSec) / rangeSec
	return pixelMin + float32(t)*(pixelMax-pixelMin)
}

// Invert implements Axis. Returns seconds as float64.
func (a *Time) Invert(pixel, pixelMin, pixelMax float32) float64 {
	minSec := timeToSeconds(a.min)
	rangeSec := timeToSeconds(a.max) - minSec
	if pixelMax == pixelMin || rangeSec == 0 {
		return minSec
	}
	t := float64(pixel-pixelMin) / float64(pixelMax-pixelMin)
	return minSec + t*rangeSec
}
