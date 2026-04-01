package lib

func CalculateOrderTokens(orderAmount float64) float64 {
	switch {
	case orderAmount < 1000:
		return orderAmount * 0.04
	case orderAmount >= 1000 && orderAmount <= 2000:
		return orderAmount * 0.05
	case orderAmount >= 3000:
		return orderAmount * 0.06
	default:
		return 0
	}
}

func CalculateStreakTokens(streakDays int, totalDaysInMonth int) float64 {
	tokens := 0.0

	// Daily token for every visit
	tokens += float64(streakDays) * 5

	// Extra token milestones
	if streakDays >= 3 {
		tokens += 5 // bonus for 3-day streak
	}
	if streakDays >= 5 {
		tokens += 5 // bonus for 5-day streak
	}

	// Monthly milestones (only once per month)
	switch {
	case totalDaysInMonth >= 15:
		tokens += 25
	case totalDaysInMonth >= 10:
		tokens += 15
	}

	return tokens
}

func CalculateDiscountFromTokens(tokens float64) float64 {
	if tokens > 100 {
		return tokens * 0.5
	}
	return 0
}
