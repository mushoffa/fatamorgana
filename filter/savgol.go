package filter

import (
	"errors"
	"math"
)

// EdgeMode selects how the filter treats the first and last window/2 samples,
// where the window extends past the data.
type EdgeMode int

const (
	EdgeInterpolate EdgeMode = iota
	EdgeNearest
	EdgeMirror
	EdgeZero
)

var (
	// ErrWindowSize is returned when the window is not a positive odd number.
	ErrWindowSize = errors.New("filter: window must be a positive odd number")

	// ErrOrder is returned when the polynomial order is negative or not smaller
	// than the window.
	ErrOrder = errors.New("filter: polynomial order must be >= 0 and < window")

	// ErrDeriv is returned when the derivative order is negative or larger than
	// the polynomial order.
	ErrDeriv = errors.New("filter: derivative order must be >= 0 and <= polynomial order")

	// ErrDelta is returned for a negative or NaN sample spacing.
	ErrDelta = errors.New("filter: delta must not be negative")

	// ErrEdgeMode is returned for an unknown EdgeMode.
	ErrEdgeMode = errors.New("filter: unknown edge mode")

	// ErrDataTooShort is returned by EdgeInterpolate when len(data) < window.
	ErrDataTooShort = errors.New("filter: data is shorter than the window")

	// ErrIllConditioned is returned when the least-squares system cannot be
	// solved, which happens for very high polynomial orders. Lower the order.
	ErrIllConditioned = errors.New("filter: least-squares system is ill-conditioned; lower the polynomial order")
)

// Options configures SavitzkyGolay. The zero value smooths (Deriv 0) with unit
// sample spacing and EdgeInterpolate.
type Options struct {
	// Deriv is the derivative order to compute. 0 smooths the data; 1 returns
	// the smoothed first derivative, and so on. It must be <= the polynomial order.
	Deriv int

	// Delta is the spacing between samples, used to scale derivatives so they
	// are expressed per unit of x. Zero means 1. It has no effect when Deriv is 0.
	Delta float64

	// Edge selects the edge handling. The zero value is EdgeInterpolate.
	Edge EdgeMode
}

func Coefficients(window, order, deriv int) ([]float64, error) {
	p, err := newPlan(window, order, deriv)
	if err != nil {
		return nil, err
	}
	return append([]float64(nil), p.coeffs...), nil
}

func Smooth(data []float64, window, order int) ([]float64, error) {
	return SavitzkyGolay(data, window, order, Options{})
}

func SmoothInts(data []int, window, order int) ([]float64, error) {
	f := make([]float64, len(data))
	for i, v := range data {
		f[i] = float64(v)
	}
	return Smooth(f, window, order)
}

func SavitzkyGolay(data []float64, window, order int, opt Options) ([]float64, error) {
	p, err := newPlan(window, order, opt.Deriv)
	if err != nil {
		return nil, err
	}

	delta := opt.Delta
	switch {
	case delta < 0 || math.IsNaN(delta):
		return nil, ErrDelta
	case delta == 0:
		delta = 1
	}

	if opt.Edge < EdgeInterpolate || opt.Edge > EdgeZero {
		return nil, ErrEdgeMode
	}

	n := len(data)
	out := make([]float64, n)
	if n == 0 {
		return out, nil
	}
	if opt.Edge == EdgeInterpolate && n < window {
		return nil, ErrDataTooShort
	}

	norm := math.Pow(delta, float64(opt.Deriv))
	h := p.half

	for i := 0; i < n; i++ {
		var s float64

		switch {
		case i >= h && i < n-h:
			// Window fully inside the data.
			base := i - h
			for k, c := range p.coeffs {
				s += c * data[base+k]
			}
			s /= norm

		case opt.Edge == EdgeInterpolate:
			s = p.edgeValue(data, i) / norm

		default:
			for k, c := range p.coeffs {
				s += c * sample(data, i-h+k, opt.Edge)
			}
			s /= norm
		}

		out[i] = s
	}

	return out, nil
}

// plan holds the precomputed least-squares solution for one window/order/deriv.
type plan struct {
	order, deriv int
	half         int
	scale        float64     // window half-width used to normalize positions to [-1, 1]
	fit          [][]float64 // (order+1) x window: polynomial coefficients from samples
	coeffs       []float64   // convolution coefficients at the window center
}

func newPlan(window, order, deriv int) (*plan, error) {
	if window < 1 || window%2 == 0 {
		return nil, ErrWindowSize
	}
	if order < 0 || order >= window {
		return nil, ErrOrder
	}
	if deriv < 0 || deriv > order {
		return nil, ErrDeriv
	}

	half := window / 2
	scale := float64(half)
	if half == 0 {
		scale = 1
	}
	n := order + 1

	// powers[k][i] = t_k^i, with t_k = (k-half)/scale in [-1, 1]. Normalizing the
	// positions keeps the normal equations well conditioned.
	powers := make([][]float64, window)
	for k := range powers {
		t := float64(k-half) / scale
		powers[k] = make([]float64, 2*order+1)
		powers[k][0] = 1
		for i := 1; i < len(powers[k]); i++ {
			powers[k][i] = powers[k][i-1] * t
		}
	}

	// Normal equations matrix M = A^T A, where A[k][j] = t_k^j.
	m := make([][]float64, n)
	for i := range m {
		m[i] = make([]float64, n)
		for j := range m[i] {
			for k := 0; k < window; k++ {
				m[i][j] += powers[k][i+j]
			}
		}
	}

	inv, err := invert(m)
	if err != nil {
		return nil, err
	}

	// fit = (A^T A)^-1 A^T maps samples to polynomial coefficients.
	fit := make([][]float64, n)
	for j := range fit {
		fit[j] = make([]float64, window)
		for k := 0; k < window; k++ {
			for i := 0; i < n; i++ {
				fit[j][k] += inv[j][i] * powers[k][i]
			}
		}
	}

	// d-th derivative at the center (t = 0) is d! * a_d / scale^d.
	factor := factorial(deriv) / math.Pow(scale, float64(deriv))
	coeffs := make([]float64, window)
	for k := range coeffs {
		coeffs[k] = factor * fit[deriv][k]
	}

	return &plan{
		order:  order,
		deriv:  deriv,
		half:   half,
		scale:  scale,
		fit:    fit,
		coeffs: coeffs,
	}, nil
}

// edgeValue evaluates the deriv-th derivative of the polynomial fitted to the
// first (or last) window of data at sample i. Used for EdgeInterpolate, which
// guarantees len(data) >= window. The result is not yet divided by delta^deriv.
func (p *plan) edgeValue(data []float64, i int) float64 {
	window := len(p.coeffs)

	start := 0
	if i >= len(data)-p.half {
		start = len(data) - window
	}

	t := float64(i-start-p.half) / p.scale

	var sum float64
	for j := p.deriv; j <= p.order; j++ {
		var a float64
		for k := 0; k < window; k++ {
			a += p.fit[j][k] * data[start+k]
		}
		sum += a * fallingFactorial(j, p.deriv) * math.Pow(t, float64(j-p.deriv))
	}

	return sum / math.Pow(p.scale, float64(p.deriv))
}

// sample returns data[i], extending the data past its ends according to mode.
func sample(data []float64, i int, mode EdgeMode) float64 {
	n := len(data)
	if i >= 0 && i < n {
		return data[i]
	}

	switch mode {
	case EdgeNearest:
		if i < 0 {
			return data[0]
		}
		return data[n-1]
	case EdgeMirror:
		return data[reflect(i, n)]
	default: // EdgeZero
		return 0
	}
}

// reflect maps any index onto [0, n) by reflecting about the first and last
// element without repeating them.
func reflect(i, n int) int {
	if n == 1 {
		return 0
	}

	period := 2 * (n - 1)
	i %= period
	if i < 0 {
		i += period
	}
	if i >= n {
		i = period - i
	}
	return i
}

func factorial(n int) float64 {
	f := 1.0
	for i := 2; i <= n; i++ {
		f *= float64(i)
	}
	return f
}

// fallingFactorial returns j! / (j-d)!.
func fallingFactorial(j, d int) float64 {
	f := 1.0
	for r := 0; r < d; r++ {
		f *= float64(j - r)
	}
	return f
}

// invert returns the inverse of a small square matrix using Gauss-Jordan
// elimination with partial pivoting.
func invert(m [][]float64) ([][]float64, error) {
	n := len(m)

	a := make([][]float64, n)
	for i := range a {
		a[i] = make([]float64, 2*n)
		copy(a[i], m[i])
		a[i][n+i] = 1
	}

	for col := 0; col < n; col++ {
		pivot := col
		for r := col + 1; r < n; r++ {
			if math.Abs(a[r][col]) > math.Abs(a[pivot][col]) {
				pivot = r
			}
		}
		if math.Abs(a[pivot][col]) < 1e-12 {
			return nil, ErrIllConditioned
		}
		a[col], a[pivot] = a[pivot], a[col]

		p := a[col][col]
		for j := range a[col] {
			a[col][j] /= p
		}

		for r := 0; r < n; r++ {
			if r == col {
				continue
			}
			f := a[r][col]
			if f == 0 {
				continue
			}
			for j := range a[r] {
				a[r][j] -= f * a[col][j]
			}
		}
	}

	inv := make([][]float64, n)
	for i := range inv {
		inv[i] = a[i][n:]
	}
	return inv, nil
}
