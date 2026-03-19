package ui

import "time"

type AnimState struct {
	Tick     int
	MaxTicks int
	Done     bool
	Interval time.Duration
}

func NewAnimState(durationMs int, intervalMs int) AnimState {
	interval := time.Duration(intervalMs) * time.Millisecond
	maxTicks := durationMs / intervalMs
	return AnimState{Tick: 0, MaxTicks: maxTicks, Interval: interval}
}

func (a *AnimState) Advance() {
	if a.Tick < a.MaxTicks {
		a.Tick++
	}
	if a.Tick >= a.MaxTicks {
		a.Done = true
	}
}

func (a *AnimState) Progress() float64 {
	if a.MaxTicks == 0 {
		return 1.0
	}
	p := float64(a.Tick) / float64(a.MaxTicks)
	if p > 1.0 {
		return 1.0
	}
	return p
}

func CounterAnimation(target int, progress float64) int {
	return int(float64(target) * progress)
}

func TypewriterAnimation(text string, progress float64) string {
	chars := int(float64(len(text)) * progress)
	if chars > len(text) {
		chars = len(text)
	}
	return text[:chars]
}

func HeatmapAnimation(totalCells int, progress float64) int {
	return int(float64(totalCells) * progress)
}
