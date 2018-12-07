package vault

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-test/deep"
)

// TestRequestCounterStorage validates the code that writes and reads
// request counters to and from the barrier.
func TestRequestCounterStorage(t *testing.T) {
	parseTime := func(format, timeval string) time.Time {
		tm, err := time.Parse(format, timeval)
		if err != nil {
			t.Fatalf("Error parsing time '%s': %v", timeval, err)
		}
		return tm
	}

	c, _, _ := TestCoreUnsealed(t)
	december2018 := parseTime(time.RFC3339, "2018-12-05T09:44:12-05:00")
	decemberRequests := uint64(555)

	atomic.StoreUint64(&c.counters.requests, decemberRequests)
	err := c.saveCurrentRequestCounters(context.Background(), december2018)
	if err != nil {
		t.Fatal(err)
	}

	atomic.StoreUint64(&c.counters.requests, 0)
	err = c.loadCurrentRequestCounters(context.Background(), december2018)
	if err != nil {
		t.Fatal(err)
	}

	if got := atomic.LoadUint64(&c.counters.requests); got != decemberRequests {
		t.Fatalf("expected=%d, got=%d", decemberRequests, got)
	}

	all, err := c.loadAllRequestCounters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	decemberStartTime := parseTime(requestCounterDatePathFormat, december2018.Format(requestCounterDatePathFormat))
	expected2018 := AllRequestCounters{
		Dated: []DatedRequestCounter{
			DatedRequestCounter{decemberStartTime, RequestCounter{Total: &decemberRequests}},
		},
	}
	if diff := deep.Equal(*all, expected2018); len(diff) != 0 {
		t.Errorf("Expected=%v, got=%v, diff=%v", expected2018, *all, diff)
	}

	january2019 := parseTime(time.RFC3339, "2019-01-02T08:21:11-05:00")
	januaryRequests := uint64(333)
	januaryStartTime := parseTime(requestCounterDatePathFormat, january2019.Format(requestCounterDatePathFormat))

	atomic.StoreUint64(&c.counters.requests, januaryRequests)
	err = c.saveCurrentRequestCounters(context.Background(), january2019)
	if err != nil {
		t.Fatal(err)
	}

	all, err = c.loadAllRequestCounters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	expected2019 := expected2018
	expected2019.Dated = append(expected2019.Dated,
		DatedRequestCounter{januaryStartTime, RequestCounter{&januaryRequests}})
	if diff := deep.Equal(*all, expected2019); len(diff) != 0 {
		t.Errorf("Expected=%v, got=%v, diff=%v", expected2019, *all, diff)
	}
}

func TestRequestCounterReset(t *testing.T) {
	parseTime := func(format, timeval string) time.Time {
		tm, err := time.Parse(format, timeval)
		if err != nil {
			t.Fatalf("Error parsing time '%s': %v", timeval, err)
		}
		return tm
	}

	c, _, _ := TestCoreUnsealed(t)

	december2018 := parseTime(time.RFC3339, "2018-12-05T09:44:12-05:00")
	decemberRequests := uint64(555)
	atomic.StoreUint64(&c.counters.requests, decemberRequests)
	err := c.saveCurrentRequestCounters(context.Background(), december2018)
	if err != nil {
		t.Fatal(err)
	}

	// Now we've written something in December.  Let's write again in December
	// and make sure nothing got reset.
	atomic.AddUint64(&c.counters.requests, 1)
	err = c.saveCurrentRequestCounters(context.Background(), december2018)
	if err != nil {
		t.Fatal(err)
	}

	if got := atomic.LoadUint64(&c.counters.requests); got != decemberRequests+1 {
		t.Errorf("expected=%d, got=%d", decemberRequests+1, got)
	}

	// Now it's January, saving counters should reset them.
	january2019 := parseTime(time.RFC3339, "2019-01-02T08:21:11-05:00")
	err = c.saveCurrentRequestCounters(context.Background(), january2019)
	if err != nil {
		t.Fatal(err)
	}

	if got := atomic.LoadUint64(&c.counters.requests); got != 0 {
		t.Errorf("expected=%d, got=%d", 0, got)
	}

	all, err := c.loadAllRequestCounters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	decemberStartTime := parseTime(requestCounterDatePathFormat, december2018.Format(requestCounterDatePathFormat))
	januaryStartTime := parseTime(requestCounterDatePathFormat, january2019.Format(requestCounterDatePathFormat))
	var expectedDecember = decemberRequests + 1
	var expectedJanuary uint64
	expected2019 := AllRequestCounters{
		Dated: []DatedRequestCounter{
			DatedRequestCounter{decemberStartTime, RequestCounter{Total: &expectedDecember}},
			DatedRequestCounter{januaryStartTime, RequestCounter{Total: &expectedJanuary}},
		},
	}
	if diff := deep.Equal(*all, expected2019); len(diff) != 0 {
		t.Errorf("Expected=%v, got=%v, diff=%v", expected2019, *all, diff)
	}
}
