package portscanner

import "time"

// SampleRecord holds a timestamped count of active ports.
type SampleRecord struct {
	At    time.Time
	Count int
}

// Sampler periodically records the number of active ports from a snapshot.
type Sampler struct {
	max     int
	records []SampleRecord
}

// NewSampler creates a Sampler that retains up to max records.
func NewSampler(max int) *Sampler {
	if max <= 0 {
		max = 60
	}
	return &Sampler{max: max}
}

// Record appends a new sample derived from the given snapshot.
func (s *Sampler) Record(snap *Snapshot, now time.Time) {
	rec := SampleRecord{
		At:    now,
		Count: len(snap.Entries()),
	}
	s.records = append(s.records, rec)
	if len(s.records) > s.max {
		s.records = s.records[len(s.records)-s.max:]
	}
}

// All returns a copy of all retained samples, oldest first.
func (s *Sampler) All() []SampleRecord {
	out := make([]SampleRecord, len(s.records))
	copy(out, s.records)
	return out
}

// Len returns the number of retained samples.
func (s *Sampler) Len() int { return len(s.records) }

// Latest returns the most recent sample and true, or zero value and false.
func (s *Sampler) Latest() (SampleRecord, bool) {
	if len(s.records) == 0 {
		return SampleRecord{}, false
	}
	return s.records[len(s.records)-1], true
}
