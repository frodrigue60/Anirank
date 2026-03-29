package utils

import "strings"

func levenshteinDistance(s1, s2 string) int {
	lenS1 := len(s1)
	lenS2 := len(s2)

	if lenS1 == 0 {
		return lenS2
	}
	if lenS2 == 0 {
		return lenS1
	}

	r1 := []rune(s1)
	r2 := []rune(s2)
	lenR1 := len(r1)
	lenR2 := len(r2)

	d := make([][]int, lenR1+1)
	for i := range d {
		d[i] = make([]int, lenR2+1)
	}

	for i := 0; i <= lenR1; i++ {
		d[i][0] = i
	}
	for j := 0; j <= lenR2; j++ {
		d[0][j] = j
	}

	for i := 1; i <= lenR1; i++ {
		for j := 1; j <= lenR2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			min1 := d[i-1][j] + 1
			min2 := d[i][j-1] + 1
			min3 := d[i-1][j-1] + cost
			
			min := min1
			if min2 < min {
				min = min2
			}
			if min3 < min {
				min = min3
			}
			d[i][j] = min
		}
	}
	return d[lenR1][lenR2]
}

// NameSimilarity calculates similarity percentage 0.0 to 1.0 (1.0 = exact match)
// It ignores case and leading/trailing whitespace
func NameSimilarity(a, b string) float64 {
	aClean := strings.ToLower(strings.TrimSpace(a))
	bClean := strings.ToLower(strings.TrimSpace(b))
	
	if aClean == bClean {
		return 1.0
	}
	
	dist := levenshteinDistance(aClean, bClean)
	
	r1 := []rune(aClean)
	r2 := []rune(bClean)
	
	maxLen := len(r1)
	if len(r2) > maxLen {
		maxLen = len(r2)
	}
	
	if maxLen == 0 {
		return 1.0
	}
	
	return 1.0 - (float64(dist) / float64(maxLen))
}
