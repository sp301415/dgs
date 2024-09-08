package dgs

import (
	"math"
	"math/big"
)

const (
	// BaseLog is the log of number of the base sampler.
	// Increasing this value will result in faster sampling, but with more memory usage.
	BaseLog = 8

	// PrecLog is the log of the precision of the sampler.
	// Should be a multiple of BaseLog.
	PrecLog = 30

	// sampleDepth is ceil(PrecLog / BaseLog).
	sampleDepth = (PrecLog + BaseLog - 1) / BaseLog

	// Eta is the smoothing parameter of Z.
	Eta = 6

	// level is the number of combining depth for convolution.
	level = 3
)

// ConvolutionSampler is a Discrete Gaussian sampler based on Convolution Theorem.
// Recommended for large and varying center and sigma.
type ConvolutionSampler struct {
	// BaseSamplers is a slice of base samplers.
	BaseSamplers [1 << BaseLog]*ReverseCDTSampler

	// z[0] = 0
	z    [level + 1]int64
	sMax float64
	sBar float64
}

// NewConvolutionSampler creates a new ConvolutionSampler.
func NewConvolutionSampler() *ConvolutionSampler {
	z := [level + 1]int64{}
	s0 := 34.0

	sMax := s0
	for i := 1; i <= level; i++ {
		z1 := int64(math.Floor(sMax / (math.Sqrt2 * Eta)))
		z2 := max(z1-1, 1)

		z[i] = z1
		sMax = math.Sqrt(float64(z1*z1+z2*z2) * sMax * sMax)
	}

	baseSamplers := [1 << BaseLog]*ReverseCDTSampler{}
	for i := 0; i < 1<<BaseLog; i++ {
		c := float64(i) / (1 << BaseLog)
		baseSamplers[i] = NewReverseCDTSampler(c, s0)
	}

	s0Big := big.NewFloat(s0).SetPrec(128)
	sBarBig := big.NewFloat(0).SetPrec(128)
	basePowBig := big.NewFloat(0).SetPrec(128)
	oneBig := big.NewFloat(1).SetPrec(128)
	for i := 0; i < sampleDepth; i++ {
		basePowBig.SetUint64(1 << (2 * i * BaseLog)) // b^2i
		basePowBig.Quo(oneBig, basePowBig)           // b^-2i
		sBarBig.Add(sBarBig, basePowBig)
	}
	sBarBig.Sqrt(sBarBig)
	sBarBig.Mul(sBarBig, s0Big)
	sBar, _ := sBarBig.Float64()

	return &ConvolutionSampler{
		BaseSamplers: baseSamplers,

		z:    z,
		sMax: sMax,
		sBar: sBar,
	}
}

func (s *ConvolutionSampler) sampleI(i int) int64 {
	if i == 0 {
		return s.BaseSamplers[0].Sample()
	}
	x1 := s.sampleI(i - 1)
	x2 := s.sampleI(i - 1)
	return s.z[i]*x1 + max(1, s.z[i]-1)*x2
}

func (s *ConvolutionSampler) sampleC(c int64) int64 {
	const mask = (1 << BaseLog) - 1
	var r int64
	for i := 0; i < sampleDepth; i++ {
		r = s.BaseSamplers[mask&c].Sample()
		if mask&c > 0 && c < 0 {
			r -= 1
		}
		c /= 1 << BaseLog
		c += r
	}
	return c
}

// Sample samples a Discrete Gaussian.
// sigma must be larger than 4sqrt(2)*Eta ~ 36, and smaller than s[max] ~ 2^40.
func (s *ConvolutionSampler) Sample(center, sigma float64) int64 {
	x := s.sampleI(level)
	K := math.Sqrt(sigma*sigma-s.sBar*s.sBar) / s.sMax
	cc := center + K*float64(x)

	ccFrac := cc - math.Floor(cc)
	ccInt := int64(cc - ccFrac)
	ccFrac64 := uint64(ccFrac * (1 << 53))

	// Randomized rounding
	const LowPrecLog = 53 - PrecLog
	ccFrac64Hi := int64(ccFrac64 >> LowPrecLog)
	ccFrac64Lo := ccFrac64 & (1<<LowPrecLog - 1)
	ccFrac64Hi += int64(s.BaseSamplers[0].UniformSampler.SampleBernoulli(float64(ccFrac64Lo) / (1 << LowPrecLog)))

	return s.sampleC(ccFrac64Hi) + ccInt
}
