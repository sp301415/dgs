package dgs_test

import (
	"math"
	"testing"

	"github.com/sp301415/dgs"
)

const (
	N = 1024
)

func meanStdDev(v []float64) (mean, stdDev float64) {
	sum := 0.0
	for _, x := range v {
		sum += x
	}

	mean = sum / float64(len(v))

	variance := 0.0
	for _, x := range v {
		variance += (x - mean) * (x - mean)
	}
	stdDev = math.Sqrt(variance / float64(len(v)))

	return
}

func meanStdDevSampleBound(meanSample, stdDevSample float64) (meanBound, stdDevBound float64) {
	k := 3.29 // From the GLITCH test suite
	meanBound = meanSample + k*stdDevSample/math.Sqrt(N)
	stdDevBound = stdDevSample + k*stdDevSample/math.Sqrt(2*(N-1))
	return
}

func TestReverseCDT(t *testing.T) {
	mean := 0.0
	sigma := 4.0 * math.Sqrt(2*math.Pi)

	s := dgs.NewReverseCDTSampler(mean, sigma)
	samples := make([]float64, N)
	for i := range samples {
		samples[i] = float64(s.Sample())
	}

	meanSampled, sigmaSampled := meanStdDev(samples)
	meanBound, sigmaBound := meanStdDevSampleBound(meanSampled, sigmaSampled)

	if math.Abs(meanSampled) > meanBound {
		t.Errorf("mean: expected %v, got %v", mean, meanSampled)
	}
	if math.Abs(sigmaSampled) > sigmaBound {
		t.Errorf("sigma: expected %v, got %v", sigma, sigmaSampled)
	}
}

func TestReverseCDTVarCenter(t *testing.T) {
	mean := 0.7
	sigma := 4.0 * math.Sqrt(2*math.Pi)

	s := dgs.NewReverseCDTVarCenterSampler(sigma)
	samples := make([]float64, N)
	for i := range samples {
		samples[i] = float64(s.Sample(mean))
	}

	meanSampled, sigmaSampled := meanStdDev(samples)
	meanBound, sigmaBound := meanStdDevSampleBound(meanSampled, sigmaSampled)

	if math.Abs(meanSampled) > meanBound {
		t.Errorf("mean: expected %v, got %v", mean, meanSampled)
	}
	if math.Abs(sigmaSampled) > sigmaBound {
		t.Errorf("sigma: expected %v, got %v", sigma, sigmaSampled)
	}
}

func TestConvolution(t *testing.T) {
	mean := 100.7
	sigma := 32 * math.Sqrt(2*math.Pi)

	s := dgs.NewConvolutionSampler(1 << 8)
	samples := make([]float64, N)
	for i := range samples {
		samples[i] = float64(s.Sample(mean, sigma))
	}

	meanSampled, sigmaSampled := meanStdDev(samples)
	meanBound, sigmaBound := meanStdDevSampleBound(meanSampled, sigmaSampled)

	if math.Abs(meanSampled) > meanBound {
		t.Errorf("mean: expected %v, got %v", mean, meanSampled)
	}
	if math.Abs(sigmaSampled) > sigmaBound {
		t.Errorf("sigma: expected %v, got %v", sigma, sigmaSampled)
	}
}
