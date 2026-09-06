package simulation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
)

type LabFuzzOptions struct {
	Seed                      uint64
	Start                     uint64
	Cases, MaxFaults, Workers int
	TwoSided, StopOnFirst     bool
	Property                  string
}
type LabFuzzSummary struct {
	Version                         string
	Options                         LabFuzzOptions
	Cases, FailingCases, Violations int
	UniqueOriginalFailures          int
	Evaluations                     map[string]int
	Failures                        []LabFuzzArtifact
}

// Bounded ordered batches allow parallel execution without retaining a campaign
// of reports. Only the coordinator selects failures or performs minimization.
// StopOnFirst counts exactly the prefix through the lowest failing case.
func RunLabFuzz(ctx context.Context, o LabFuzzOptions) (LabFuzzSummary, error) {
	return runLabFuzz(ctx, o, EvaluateLabFuzz)
}

type labFuzzEvaluator func(context.Context, LabFuzzCase, string) ([]LabFuzzViolation, map[string]int, error)

func runLabFuzz(ctx context.Context, o LabFuzzOptions, evaluate labFuzzEvaluator) (LabFuzzSummary, error) {
	out := LabFuzzSummary{Version: LabFuzzVersion, Options: o, Evaluations: map[string]int{}}
	if err := o.Validate(); err != nil {
		return out, err
	}
	seen := map[string]bool{}
	originalSeen := map[[32]byte]bool{}
	type result struct {
		c        LabFuzzCase
		failures []LabFuzzViolation
		counts   map[string]int
		err      error
	}
	for base := 0; base < o.Cases; base += o.Workers {
		size := min(o.Workers, o.Cases-base)
		batch := make([]result, size)
		var wg sync.WaitGroup
		for i := range batch {
			wg.Add(1)
			go func() {
				defer wg.Done()
				x := &batch[i]
				x.c, x.err = GenerateLabFuzz(o.Seed, o.Start+uint64(base+i), o.MaxFaults, o.TwoSided)
				if x.err == nil {
					x.failures, x.counts, x.err = evaluate(ctx, x.c, o.Property)
				}
			}()
		}
		wg.Wait()
		for _, x := range batch {
			if x.err != nil {
				return out, fmt.Errorf("case %d: %w", x.c.Index, x.err)
			}
			out.Cases++
			for p, n := range x.counts {
				out.Evaluations[p] += n
			}
			if len(x.failures) > 0 {
				out.FailingCases++
			}
			out.Violations += len(x.failures)
			for _, f := range x.failures {
				key := f.Property + "/" + f.Code + "/" + x.c.Relation
				data, err := json.Marshal(struct {
					Identity string
					Worlds   []LabScenario
				}{key, x.c.Worlds})
				if err != nil {
					return out, err
				}
				digest := sha256.Sum256(data)
				if !originalSeen[digest] {
					originalSeen[digest] = true
					out.UniqueOriginalFailures++
				}
				if seen[key] {
					continue
				}
				seen[key] = true
				a, err := minimizeLabFuzz(ctx, x.c, f, evaluate)
				if err != nil {
					return out, err
				}
				out.Failures = append(out.Failures, a)
			}
			if len(x.failures) > 0 && o.StopOnFirst {
				return out, nil
			}
		}
	}
	return out, nil
}

func (o LabFuzzOptions) Validate() error {
	if o.Cases < 1 || o.Cases > 10000000 || o.Workers < 1 || o.Workers > 64 || o.MaxFaults < 0 || o.MaxFaults > 12 || uint64(o.Cases-1) > ^uint64(0)-o.Start {
		return fmt.Errorf("invalid campaign bounds")
	}
	if o.Property != "" {
		found := false
		for _, p := range LabFuzzProperties() {
			found = found || p.Name == o.Property
		}
		if !found {
			return fmt.Errorf("unknown property %q", o.Property)
		}
	}
	return nil
}
