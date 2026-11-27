package stamp

import "time"

type Options struct {
	Now func() time.Time
}

func Year(o Options) int {
	now := time.Now
	if o.Now != nil {
		now = o.Now
	}
	return now().Year()
}
