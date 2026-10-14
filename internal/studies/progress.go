package studies

// Progress is the share of a subject already covered, from 0 to 1: reviewed
// topics count fully and topics being studied count half.
func Progress(topics []Topic) float64 {
	if len(topics) == 0 {
		return 0
	}
	var points float64
	for _, t := range topics {
		switch t.Status {
		case StatusReviewed:
			points += 1
		case StatusStudying:
			points += 0.5
		}
	}
	return points / float64(len(topics))
}
