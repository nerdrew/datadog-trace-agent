package sampler

// InternalState exposes all the main internal settings of the scope sampler
type InternalState struct {
	Offset      float64
	Slope       float64
	Cardinality int64
	InTPS       float64
	OutTPS      float64
	MaxTPS      float64
}

// GetState collects and return internal statistics and coefficients for indication purposes
func (s *ScoreSampler) GetState() InternalState {
	return InternalState{
		s.signatureScoreOffset,
		s.signatureScoreSlope,
		s.Backend.GetCardinality(),
		s.Backend.GetTotalScore(),
		s.Backend.GetSampledScore(),
		s.maxTPS,
	}
}
