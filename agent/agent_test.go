package main

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	log "github.com/cihub/seelog"

	"github.com/DataDog/datadog-trace-agent/config"
	"github.com/DataDog/datadog-trace-agent/fixtures"
)

func TestWatchdog(t *testing.T) {
	if testing.Short() {
		return
	}

	conf := config.NewDefaultAgentConfig()
	conf.APIKey = "apikey_2"
	conf.MaxMemory = 1e7
	conf.WatchdogInterval = time.Millisecond

	// save the global mux aside, we don't want to break other tests
	defaultMux := http.DefaultServeMux
	http.DefaultServeMux = http.NewServeMux()

	agent := NewAgent(conf)

	defer func() {
		close(agent.exit)
		// We need to manually close the receiver as the Run() func
		// should have been broken and interrupted by the watchdog panic
		close(agent.Receiver.exit)
		// we need to wait more than on second (time for StoppableListener.Accept
		// to acknowledge the connection has been closed)
		time.Sleep(2 * time.Second)
		http.DefaultServeMux = defaultMux
	}()

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case string:
				if strings.HasPrefix(v, "exceeded max memory") {
					t.Logf("watchdog worked, trapped the right error: %s", v)
					runtime.GC() // make sure we clean up after allocating all this
					return
				}
			}
			t.Fatalf("unexpected error: %v", r)
		}
	}()

	// allocating a lot of memory
	buf := make([]byte, 2*int64(conf.MaxMemory))
	buf[0] = 1
	buf[len(buf)-1] = 1

	// override the default die, else our test would stop, use a plain panic() instead
	agent.die = func(format string, args ...interface{}) {
		panic(fmt.Sprintf(format, args...))
	}

	// after some time, the watchdog should kill this
	agent.Run()

	// without this. runtime could be smart and free memory before we Run()
	buf[0] = 2
	buf[len(buf)-1] = 2
}

func BenchmarkAgentTraceProcessing(b *testing.B) {
	c := config.NewDefaultAgentConfig()
	c.APIKey = "test"

	runTraceProcessingBenchmark(b, c)
}

func BenchmarkAgentTraceProcessingWithFiltering(b *testing.B) {
	c := config.NewDefaultAgentConfig()
	c.APIKey = "test"
	c.Ignore["resource"] = []string{"[0-9]{3}", "foobar", "G.T [a-z]+", "[^123]+_baz"}

	runTraceProcessingBenchmark(b, c)
}

// worst case scenario: spans are tested against multiple rules without any match.
// this means we won't compesate the overhead of filtering by dropping traces
func BenchmarkAgentTraceProcessingWithWorstCaseFiltering(b *testing.B) {
	c := config.NewDefaultAgentConfig()
	c.APIKey = "test"
	c.Ignore["resource"] = []string{"[0-9]{3}", "foobar", "aaaaa?aaaa", "[^123]+_baz"}

	runTraceProcessingBenchmark(b, c)
}

func runTraceProcessingBenchmark(b *testing.B, c *config.AgentConfig) {
	agent := NewAgent(c)
	log.UseLogger(log.Disabled)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		agent.Process(fixtures.RandomTrace(10, 8))
	}
}

func BenchmarkWatchdog(b *testing.B) {
	conf := config.NewDefaultAgentConfig()
	conf.APIKey = "apikey_2"
	agent := NewAgent(conf)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		agent.watchdog()
	}
}
