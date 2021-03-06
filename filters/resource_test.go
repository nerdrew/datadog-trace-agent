package filters

import (
	"testing"

	"github.com/DataDog/datadog-trace-agent/config"
	"github.com/DataDog/datadog-trace-agent/fixtures"
	"github.com/DataDog/datadog-trace-agent/model"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		filter      string
		resource    string
		expectation bool
	}{
		{"/foo/bar", "/foo/bar", false},
		{"/foo/b.r", "/foo/bar", false},
		{"[0-9]+", "/abcde", true},
		{"[0-9]+", "/abcde123", false},
		{"\\(foobar\\)", "(foobar)", false},
		{"\\(foobar\\)", "(bar)", true},
	}

	for _, test := range tests {
		span := newTestSpan(test.resource)
		filter := newTestFilter(test.filter)

		assert.Equal(t, test.expectation, filter.Keep(span))
	}
}

// a filter instantiated with malformed expressions should let anything pass
func TestRegexCompilationFailure(t *testing.T) {
	filter := newTestFilter("[123", "]123", "{6}")

	for i := 0; i < 100; i++ {
		span := fixtures.RandomSpan()
		assert.True(t, filter.Keep(&span))
	}
}

func TestRegexEscaping(t *testing.T) {
	span := newTestSpan("[123")

	filter := newTestFilter("[123")
	assert.True(t, filter.Keep(span))

	filter = newTestFilter("\\[123")
	assert.False(t, filter.Keep(span))
}

func TestMultipleEntries(t *testing.T) {
	filter := newTestFilter("ABC+", "W+")

	span := newTestSpan("ABCCCC")
	assert.False(t, filter.Keep(span))

	span = newTestSpan("WWW")
	assert.False(t, filter.Keep(span))
}

func newTestFilter(blacklist ...string) Filter {
	c := config.NewDefaultAgentConfig()
	c.Ignore["resource"] = blacklist

	return newResourceFilter(c)
}

func newTestSpan(resource string) *model.Span {
	span := fixtures.RandomSpan()
	span.Resource = resource
	return &span
}
