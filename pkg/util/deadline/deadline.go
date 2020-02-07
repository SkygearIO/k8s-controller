package deadline

import "time"

type Deadline struct {
	t *time.Time
}

func (d *Deadline) Set(t time.Time) {
	if d.t == nil || t.Before(*d.t) {
		d.t = &t
	}
}

func (d *Deadline) Duration(from time.Time) time.Duration {
	if d.t == nil {
		// no deadline
		return 0
	} else {
		duration := d.t.Sub(from)
		if duration <= 0 {
			// has expired deadline, use epsilon value
			duration = 1
		}
		return duration
	}
}
