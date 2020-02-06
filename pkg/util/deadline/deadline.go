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
		return time.Duration(0)
	} else {
		return d.t.Sub(from)
	}
}
