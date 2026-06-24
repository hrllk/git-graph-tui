package app

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func emptyDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

func shorten(v string, n int) string {
	if v == "" || len(v) <= n {
		return v
	}
	return v[:n]
}
