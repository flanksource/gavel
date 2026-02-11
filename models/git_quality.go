package models

func (c CommitAnalysis) GetQualityScore() int {
	return GetQualityScore(c)
}

func GetQualityScore(c CommitAnalysis) int {
	score := 0
	if c.CommitType != CommitTypeUnknown {
		score += 5
	}
	if c.Scope != ScopeTypeUnknown {
		score += 5
	}
	if len(c.Subject) > 80 {
		score += 5
	} else if len(c.Subject) > 50 {
		score += 20
	} else if len(c.Subject) > 25 {
		score += 10
	} else if len(c.Subject) > 10 {
		score += 5
	}

	if len(c.Body) > 200 {
		score += 20
	} else if len(c.Body) > 100 {
		score += 10
	} else if len(c.Body) > 50 {
		score += 5
	}

	if len(c.Trailers) > 0 {
		score += 5
	}
	return score

}
