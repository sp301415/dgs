package dgs_test

import (
	"math/rand"
	"testing"

	"github.com/sp301415/dgs"
	"gonum.org/v1/gonum/stat"
)

func TestReverseCDT(t *testing.T) {
	mean := (rand.Float64() * 128) - 64
	sigma := rand.Float64() * 64

	s := dgs.NewReverseCDTSampler(mean, sigma)
	samples := make([]float64, 1024)
	for i := range samples {
		samples[i] = float64(s.Sample())
	}

	meanSampled, sigmaSampled := stat.MeanStdDev(samples, nil)
	if meanSampled < mean-5 || meanSampled > mean+5 {
		t.Errorf("mean: expected %v, got %v", mean, meanSampled)
	}
	if sigmaSampled < sigma-5 || sigmaSampled > sigma+5 {
		t.Errorf("sigma: expected %v, got %v", sigma, sigmaSampled)
	}
}

func TestConvolution(t *testing.T) {
	mean := (rand.Float64() * 128) - 64
	var sigma float64
	for sigma < 36 {
		sigma = rand.Float64() * 128
	}

	s := dgs.NewConvolutionSampler()
	samples := make([]float64, 1024)
	for i := range samples {
		samples[i] = float64(s.Sample(mean, sigma))
	}

	meanSampled, sigmaSampled := stat.MeanStdDev(samples, nil)
	if meanSampled < mean-5 || meanSampled > mean+5 {
		t.Errorf("mean: expected %v, got %v", mean, meanSampled)
	}
	if sigmaSampled < sigma-5 || sigmaSampled > sigma+5 {
		t.Errorf("sigma: expected %v, got %v", sigma, sigmaSampled)
	}
}
