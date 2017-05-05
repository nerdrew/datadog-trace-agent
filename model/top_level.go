package model

const (
	// TraceMetricsTagKey is a tag key which, if set to true,
	// ensures all statistics are computed for this span.
	TraceMetricsTagKey = "datadog.trace_metrics"

	subNameTag   = "_sub_name"
	trueTagValue = "true"
)

// ComputeTopLevel updates all the spans top-level attribute.
//
// A span is considered top-level if:
// - it's a root span
// - its parent is unknown (other part of the code, distributed trace)
// - its parent belongs to another service (in that case it's a "local root"
//   being the highest ancestor of other spans belonging to this service and
//   attached to it).
func (t Trace) ComputeTopLevel() {
	// build a lookup map
	spanIDToIdx := make(map[uint64]int, len(t))
	for i, v := range t {
		spanIDToIdx[v.SpanID] = i
	}

	// iterate on each span and mark them as top-level if relevant
	for i, span := range t {
		if span.ParentID == 0 {
			continue
		}
		parentIdx, ok := spanIDToIdx[span.ParentID]
		if !ok {
			continue
		}
		if t[parentIdx].Service != span.Service {
			continue
		}
		t[i].setTopLevel(false)
	}
}

// setTopLevel sets the top-level attribute of the span.
func (s *Span) setTopLevel(topLevel bool) {
	if topLevel == true {
		if s.Meta == nil {
			return
		}
		delete(s.Meta, subNameTag)
		if len(s.Meta) == 0 {
			s.Meta = nil
		}
		return
	}
	if s.Meta == nil {
		s.Meta = make(map[string]string, 1)
	}
	s.Meta[subNameTag] = trueTagValue
}

// TopLevel returns true if span is top-level.
func (s *Span) TopLevel() bool {
	return !(s.Meta[subNameTag] == trueTagValue)
}

// SkipStats returns true if statistics should not be computed for this span.
func (s *Span) SkipStats() bool {
	if s.Meta[TraceMetricsTagKey] == trueTagValue {
		return false
	}
	return !s.TopLevel()
}
