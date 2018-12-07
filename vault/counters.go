package vault

import (
	"context"
	"sort"
	"sync/atomic"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/logical"
)

const requestCounterDatePathFormat = "2006/01"

// RequestCounter stores the state of request counters for a single unspecified period.
type RequestCounter struct {
	Total *uint64 `json:"total"`
}

type DatedRequestCounter struct {
	StartTime time.Time `json:"start_time"`
	RequestCounter
}

type AllRequestCounters struct {
	// Dated holds the request counters dating back to when the feature was first
	// introduced in this instance, ordered by time (oldest first).
	Dated []DatedRequestCounter
}

func (c *Core) loadAllRequestCounters(ctx context.Context) (*AllRequestCounters, error) {
	view := c.systemBarrierView.SubView("counters/requests/")

	datepaths, err := view.List(ctx, "")
	if err != nil {
		return nil, errwrap.Wrapf("failed to read request counters: {{err}}", err)
	}
	if datepaths == nil {
		return nil, nil
	}

	var all AllRequestCounters
	sort.Strings(datepaths)
	for _, datepath := range datepaths {
		datesubpaths, err := view.List(ctx, datepath)
		if err != nil {
			return nil, errwrap.Wrapf("failed to read request counters: {{err}}", err)
		}
		for _, datesubpath := range datesubpaths {
			fullpath := datepath + datesubpath
			counter, err := c.loadRequestCounters(ctx, fullpath)
			if err != nil {
				return nil, err
			}

			t, err := time.Parse(requestCounterDatePathFormat, fullpath)
			if err != nil {
				return nil, err
			}

			all.Dated = append(all.Dated, DatedRequestCounter{StartTime: t, RequestCounter: *counter})
		}
	}

	return &all, nil
}

func (c *Core) loadCurrentRequestCounters(ctx context.Context, now time.Time) error {
	datepath := now.Format(requestCounterDatePathFormat)
	counter, err := c.loadRequestCounters(ctx, datepath)
	if err != nil {
		return err
	}
	if counter != nil {
		atomic.StoreUint64(&c.counters.requests, *counter.Total)
	}
	return nil
}

func (c *Core) loadRequestCounters(ctx context.Context, datepath string) (*RequestCounter, error) {
	view := c.systemBarrierView.SubView("counters/requests/")

	out, err := view.Get(ctx, datepath)
	if err != nil {
		return nil, errwrap.Wrapf("failed to read request counters: {{err}}", err)
	}
	if out == nil {
		return nil, nil
	}

	newCounters := &RequestCounter{}
	err = out.DecodeJSON(newCounters)
	if err != nil {
		return nil, err
	}

	return newCounters, nil
}

func (c *Core) saveCurrentRequestCounters(ctx context.Context, now time.Time) error {
	view := c.systemBarrierView.SubView("counters/requests/")

	requests := atomic.LoadUint64(&c.counters.requests)
	localCounters := &RequestCounter{
		Total: &requests,
	}

	datepath := now.Format(requestCounterDatePathFormat)
	entry, err := logical.StorageEntryJSON(datepath, localCounters)
	if err != nil {
		return errwrap.Wrapf("failed to create request counters entry: {{err}}", err)
	}

	if err := view.Put(ctx, entry); err != nil {
		return errwrap.Wrapf("failed to save request counters: {{err}}", err)
	}

	return nil
}
