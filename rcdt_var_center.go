package dgs

import (
	"math"
)

const (
	// BaseLog is the log of number of the base sampler.
	// Increasing this value will result in faster sampling, but with more memory usage.
	BaseLog = 6

	// SampleDepth is ceil(PrecLog / BaseLog).
	SampleDepth = 5

	// hiPrecLog is the precision of the sampler.
	hiPrecLog = BaseLog * SampleDepth

	// lowPrecLog is the precision of randomized rounding.
	lowPrecLog = 53 - hiPrecLog
)

// ReverseCDTVarCenterSampler is a Discrete Gaussian sampler based on Reverse Cumulative Distribution Table.
// Recommended for varying center and fixed sigma.
//
// Note that the stdandard devation of the samples is sigma/sqrt(2pi).
type ReverseCDTVarCenterSampler struct {
	// BaseSamplers is a base uniform sampler.
	BaseSamplers [1 << BaseLog]*ReverseCDTSampler
}

// NewReverseCDTVarCenterSampler creates a new ReverseCDTVarCenterSampler.
func NewReverseCDTVarCenterSampler(sigma float64) *ReverseCDTVarCenterSampler {
	// // Fixing sigma blowup
	// bk := 0.0
	// for i := 0; i < SampleDepth; i++ {
	// 	bk += math.Pow(1<<BaseLog, -2*float64(i))
	// }
	// sigma /= math.Sqrt(bk)

	baseSamplers := [1 << BaseLog]*ReverseCDTSampler{}
	for i := 0; i < 1<<BaseLog; i++ {
		c := float64(i) / (1 << BaseLog)
		baseSamplers[i] = NewReverseCDTSampler(c, sigma)
	}

	return &ReverseCDTVarCenterSampler{
		BaseSamplers: baseSamplers,
	}
}

// Sample samples from the distribution.
func (s *ReverseCDTVarCenterSampler) Sample(center float64) int64 {
	cFrac := center - math.Floor(center)
	cInt := int64(center - cFrac)
	cFrac64 := uint64(cFrac * (1 << 53))

	cFrac64Hi := int64(cFrac64 >> lowPrecLog)
	r := s.BaseSamplers[0].UniformSampler.Sample()
	for i := lowPrecLog - 1; i >= 0; i-- {
		b := (r >> i) & 1
		cFracBit := (cFrac64 >> i) & 1
		if b > cFracBit {
			return s.sampleC(cFrac64Hi) + cInt
		}
		if b < cFracBit {
			return s.sampleC(cFrac64Hi+1) + cInt
		}
	}

	return s.sampleC(cFrac64Hi+1) + cInt
}

func (s *ReverseCDTVarCenterSampler) sampleC(c int64) int64 {
	const mask = (1 << BaseLog) - 1
	for i := 0; i < SampleDepth; i++ {
		r := s.BaseSamplers[c&mask].Sample()
		if c&mask > 0 && c < 0 {
			r -= 1
		}
		c >>= BaseLog
		c += r
	}
	return c
}
