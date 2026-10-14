package filter

import (
	"errors"
	"math"
	"sort"
)

// ErrLineOptions is returned by SegmentLines for invalid LineOptions: negative
// sizes, or ratios outside [0, 1].
var ErrLineOptions = errors.New("filter: invalid line segmentation options")

// Peak is a local maximum (or, for FindValleys, a local minimum) of a signal.
type Peak struct {
	Index int
	Value float64
	Prominence float64
}

// PeakOptions filters the candidates found by FindPeaks and FindValleys.
type PeakOptions struct {
	// MinProminence discards peaks whose prominence is below this value.
	// Zero or negative disables the filter.
	MinProminence float64

	// MinDistance is the minimum number of samples between neighboring peaks.
	// When two peaks are closer, the higher one is kept (the deeper one, for
	// valleys). Values of 1 or less disable the filter.
	MinDistance int
}

// FindPeaks returns the local maxima of data, in index order.
func FindPeaks(data []float64, opt PeakOptions) []Peak {
	var peaks []Peak

	for _, i := range localMaxima(data) {
		prom := prominence(data, i)
		if opt.MinProminence > 0 && prom < opt.MinProminence {
			continue
		}
		peaks = append(peaks, Peak{Index: i, Value: data[i], Prominence: prom})
	}

	if opt.MinDistance > 1 {
		peaks = enforceDistance(peaks, opt.MinDistance)
	}
	return peaks
}

// FindValleys returns the local minima of data, in index order. It takes the
// same options as FindPeaks. In the result, Value is the original sample value
// and Prominence is the depth of the valley.
func FindValleys(data []float64, opt PeakOptions) []Peak {
	neg := make([]float64, len(data))
	for i, v := range data {
		neg[i] = -v
	}

	valleys := FindPeaks(neg, opt)
	for i := range valleys {
		valleys[i].Value = data[valleys[i].Index]
	}
	return valleys
}

// localMaxima returns the indices of strict local maxima, using the middle of
// flat tops.
func localMaxima(data []float64) []int {
	n := len(data)
	var out []int

	i := 1
	for i < n-1 {
		if data[i-1] < data[i] {
			// Walk to the end of a possible plateau.
			j := i
			for j+1 < n && data[j+1] == data[i] {
				j++
			}
			// It is a peak only if the data falls after the plateau.
			if j+1 < n && data[j+1] < data[i] {
				out = append(out, (i+j)/2)
			}
			i = j + 1
		} else {
			i++
		}
	}
	return out
}

// prominence computes the prominence of the peak at index i.
func prominence(data []float64, i int) float64 {
	v := data[i]
	n := len(data)

	leftMin := v
	for k := i - 1; k >= 0 && data[k] <= v; k-- {
		if data[k] < leftMin {
			leftMin = data[k]
		}
	}

	rightMin := v
	for k := i + 1; k < n && data[k] <= v; k++ {
		if data[k] < rightMin {
			rightMin = data[k]
		}
	}

	return v - math.Max(leftMin, rightMin)
}

// enforceDistance removes lower peaks that are closer than distance samples to
// a higher one. peaks must be in index order, and so is the result.
func enforceDistance(peaks []Peak, distance int) []Peak {
	order := make([]int, len(peaks))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return peaks[order[a]].Value > peaks[order[b]].Value
	})

	keep := make([]bool, len(peaks))
	for i := range keep {
		keep[i] = true
	}

	for _, p := range order {
		if !keep[p] {
			continue
		}
		for k := p - 1; k >= 0 && peaks[p].Index-peaks[k].Index < distance; k-- {
			keep[k] = false
		}
		for k := p + 1; k < len(peaks) && peaks[k].Index-peaks[p].Index < distance; k++ {
			keep[k] = false
		}
	}

	var out []Peak
	for i, p := range peaks {
		if keep[i] {
			out = append(out, p)
		}
	}
	return out
}

// Band is a run of consecutive rows (or columns) of a profile, such as one text
// line. Start is inclusive and End is exclusive.
type Band struct {
	Start, End int
}

// Height returns the number of rows in the band.
func (b Band) Height() int {
	return b.End - b.Start
}

// LineOptions configures SegmentLines.
type LineOptions struct {
	SmoothWindow int
	SmoothOrder  int
	MinProminence float64
	MinLineHeight int
	TrimRatio float64
}

// DefaultLineOptions returns starting values suited to roughly 20 to 80 pixel
// text lines. Tune SmoothWindow and MinProminence for your scans.
func DefaultLineOptions() LineOptions {
	return LineOptions{
		SmoothWindow:  11,
		SmoothOrder:   3,
		MinProminence: 0.1,
		MinLineHeight: 8,
		TrimRatio:     0.05,
	}
}

func SegmentLines(profile []int, opt LineOptions) ([]Band, error) {
	if opt.SmoothWindow < 0 || opt.MinLineHeight < 0 ||
		opt.MinProminence < 0 || opt.MinProminence > 1 ||
		opt.TrimRatio < 0 || opt.TrimRatio > 1 {
		return nil, ErrLineOptions
	}

	n := len(profile)
	if n == 0 {
		return nil, nil
	}

	smooth := make([]float64, n)
	for i, v := range profile {
		smooth[i] = float64(v)
	}

	if opt.SmoothWindow > 0 {
		var err error
		// EdgeNearest works for profiles of any length and avoids the polynomial
		// overshoot that edge interpolation can add at the page border.
		smooth, err = SavitzkyGolay(smooth, opt.SmoothWindow, opt.SmoothOrder, Options{Edge: EdgeNearest})
		if err != nil {
			return nil, err
		}
	}

	maxVal := smooth[0]
	for _, v := range smooth {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal <= 0 {
		return nil, nil
	}

	minProm := opt.MinProminence * maxVal
	valleys := FindValleys(smooth, PeakOptions{
		MinProminence: minProm,
		MinDistance:   opt.MinLineHeight,
	})

	cuts := make([]int, 0, len(valleys)+2)
	cuts = append(cuts, 0)
	for _, v := range valleys {
		cuts = append(cuts, v.Index)
	}
	cuts = append(cuts, n)

	var bands []Band
	for k := 0; k+1 < len(cuts); k++ {
		s, e := cuts[k], cuts[k+1]

		peak := smooth[s]
		for i := s; i < e; i++ {
			if smooth[i] > peak {
				peak = smooth[i]
			}
		}
		// Skip margins and noise: a line must rise clearly above the background.
		if peak <= 0 || peak < minProm {
			continue
		}

		thr := opt.TrimRatio * peak
		start, end := s, e
		for start < e && smooth[start] < thr {
			start++
		}
		for end > start && smooth[end-1] < thr {
			end--
		}

		if end <= start || end-start < opt.MinLineHeight {
			continue
		}
		bands = append(bands, Band{Start: start, End: end})
	}

	return bands, nil
}
