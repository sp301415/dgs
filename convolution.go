package dgs

import (
	"math"
)

const Eta = 6.0

// ConvolutionSampler is a Discrete Gaussian sampler based on Convolution.
//
// Note that the stdandard devation of the samples is sigma/sqrt(2pi).
type ConvolutionSampler struct {
	BaseSampler *ReverseCDTVarCenterSampler

	z      []int64
	s      []float64
	sigBar float64
}

// NewConvolutionSampler creates a new ReverseCDTVarStdDevSampler.
func NewConvolutionSampler(maxSigma float64) *ConvolutionSampler {
	convolveDepth := int(math.Ceil(math.Log2(math.Log2(maxSigma))))

	z := make([]int64, convolveDepth+1)
	s := make([]float64, convolveDepth+1)
	s[0] = 4 * math.Sqrt2 * Eta
	for i := 1; i < convolveDepth+1; i++ {
		z[i] = int64(math.Floor(s[i-1] / (math.Sqrt2 * Eta)))
		s[i] = math.Sqrt(float64(z[i]*z[i]+max(1, (z[i]-1)*(z[i]-1)))) * s[i-1]
	}

	sigBar := 0.0
	for i := 0; i < SampleDepth; i++ {
		sigBar += math.Pow(1<<BaseLog, -2*float64(i))
	}
	sigBar *= s[0]

	return &ConvolutionSampler{
		BaseSampler: NewReverseCDTVarCenterSampler(s[0]),

		z:      z,
		s:      s,
		sigBar: sigBar,
	}
}

// Sample samples from the distribution.
func (s *ConvolutionSampler) Sample(center, sigma float64) int64 {
	var m int
	for m = 0; m < len(s.s); m++ {
		if s.s[m] >= sigma {
			break
		}
	}

	x := s.sampleI(m)
	K := math.Sqrt(sigma*sigma-s.sigBar*s.sigBar) / s.s[m]
	return s.BaseSampler.Sample(center + K*float64(x))
}

func (s *ConvolutionSampler) sampleI(i int) int64 {
	if i == 0 {
		return s.BaseSampler.BaseSamplers[0].Sample()
	}

	x1 := s.sampleI(i - 1)
	x2 := s.sampleI(i - 1)

	return s.z[i]*x1 + max(1, s.z[i]-1)*x2
}
