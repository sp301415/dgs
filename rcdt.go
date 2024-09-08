package dgs

import (
	"math"
	"math/big"
	"slices"

	"github.com/ALTree/bigfloat"
)

const TailCut = 6

// ReverseCDTSampler is a Discrete Gaussian sampler based on Reverse Cumulative Distribution Table.
// Recommended for fixed and small center and sigma.
type ReverseCDTSampler struct {
	// UniformSampler is a base uniform sampler.
	UniformSampler *UniformSampler

	// Table is a table of probabilities.
	Table []uint64

	// Center is the center of the distribution.
	Center float64
	// Sigma is the standard deviation of the distribution.
	Sigma float64

	// TailLow is the lower bound of the tail.
	// Currently, tail cut of 6 sigma is used.
	TailLow int64
	// TailHigh is the upper bound of the tail.
	// Currently, tail cut of 6 sigma is used.
	TailHigh int64

	// cInt is the integer part of the center.
	cInt int64
	// cFrac is the fractional part of the center.
	cFrac float64
}

// NewReverseCDTSampler creates a new ReverseCDTSampler.
func NewReverseCDTSampler(center, sigma float64) *ReverseCDTSampler {
	cFrac := center - math.Floor(center)
	cInt := int64(center - cFrac)

	tailLow := int64(math.Round(center - TailCut*sigma))
	tailHigh := int64(math.Round(center + TailCut*sigma))
	tailCount := int(tailHigh - tailLow + 1)

	// Use sufficient precision, usually 64 bits are enough.
	// But, here we use 128 bits to be safe.
	cBig := big.NewFloat(cFrac).SetPrec(128) // c
	sBig := big.NewFloat(sigma).SetPrec(128) // s

	sBigDoubleSq := big.NewFloat(0).SetPrec(128).Mul(sBig, sBig) // 2s^2
	sBigDoubleSq.Add(sBigDoubleSq, sBigDoubleSq)
	sBigDoubleSq.Neg(sBigDoubleSq)

	piBig := big.NewFloat(math.Pi).SetPrec(128) // -pi
	piBig.Neg(piBig)

	// Buffers
	xBig := big.NewFloat(0).SetPrec(128)
	logRhoBig := big.NewFloat(0).SetPrec(128)

	// Fill the Gaussian Table, without normalization.
	// We want T[i+1] - T[i] = rho(i).
	tableBig := make([]*big.Float, tailCount+1)
	tableBig[0] = big.NewFloat(0).SetPrec(128)
	for i, x := 1, tailLow; x <= tailHigh; i, x = i+1, x+1 {
		xBig.SetFloat64(float64(x))
		logRhoBig.Sub(xBig, cBig)              // x - c
		logRhoBig.Abs(logRhoBig)               // |x - c|
		logRhoBig.Mul(logRhoBig, logRhoBig)    // |x - c|^2
		logRhoBig.Quo(logRhoBig, sBigDoubleSq) // - |x - c|^2 / 2s^2

		tableBig[i] = bigfloat.Exp(logRhoBig)       // exp(- |x - c|^2 / 2s^2)
		tableBig[i].Add(tableBig[i-1], tableBig[i]) // Save the cumulative sum
	}

	// Normalize the Gaussian Table.
	Exp64Big := big.NewFloat(0).SetPrec(128).SetMantExp(big.NewFloat(1), 64) // 2^64
	sumBig := tableBig[len(tableBig)-1]
	table := make([]uint64, len(tableBig))
	for i := range tableBig {
		tableBig[i].Quo(tableBig[i], sumBig)   // Normalize
		tableBig[i].Mul(tableBig[i], Exp64Big) // Move to uint64
		table[i], _ = tableBig[i].Uint64()
	}

	return &ReverseCDTSampler{
		UniformSampler: NewUniformSampler(),

		Table: table,

		Center: center,
		Sigma:  sigma,

		TailLow:  tailLow,
		TailHigh: tailHigh,

		cInt:  cInt,
		cFrac: cFrac,
	}
}

// Sample samples one value from the distribution.
func (s *ReverseCDTSampler) Sample() int64 {
	u := s.UniformSampler.Sample()
	r, _ := slices.BinarySearch(s.Table, u)
	// Here, we ignore the (extremely rare) case when s.Table[r] = u.
	return int64(r) + s.cInt + s.TailLow
}
