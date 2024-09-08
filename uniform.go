package dgs

import (
	"crypto/rand"
	"math"
)

const BufferSize = 4096

// UniformSampler is a wrapper around a buffer of crypto/rand.
type UniformSampler struct {
	buf [BufferSize]byte
	ptr int
}

// NewUniformSampler creates a new UniformSampler.
func NewUniformSampler() *UniformSampler {
	return &UniformSampler{
		ptr: BufferSize,
	}
}

// Sample samples one uint64 value.
func (s *UniformSampler) Sample() uint64 {
	if s.ptr == BufferSize {
		_, err := rand.Read(s.buf[:])
		if err != nil {
			panic(err)
		}
		s.ptr = 0
	}

	var u uint64
	u |= uint64(s.buf[s.ptr+0]) << 0
	u |= uint64(s.buf[s.ptr+1]) << 8
	u |= uint64(s.buf[s.ptr+2]) << 16
	u |= uint64(s.buf[s.ptr+3]) << 24
	u |= uint64(s.buf[s.ptr+4]) << 32
	u |= uint64(s.buf[s.ptr+5]) << 40
	u |= uint64(s.buf[s.ptr+6]) << 48
	u |= uint64(s.buf[s.ptr+7]) << 56
	s.ptr += 8

	return u
}

// SampleBit samples random bit.
func (s *UniformSampler) SampleBit() uint64 {
	// TODO: Make optimal use of buffer.
	// Now, it uses 8 bytes to sample only one bit.
	return s.Sample() & 1
}

// SampleBernoulli samples random bit with given probability.
// p should be in [0, 1].
func (s *UniformSampler) SampleBernoulli(p float64) uint64 {
	p64 := uint64(math.Round(math.Ldexp(p, 64)))
	if s.Sample() < p64 {
		return 1
	}
	return 0
}
