package domain_test

import (
	"anirank/api/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterHighestBadges(t *testing.T) {
	levelType := "level"
	ratingType := "ratings"

	badges := []domain.Badge{
		{Name: "Bronze IV", RequirementType: &levelType, RequirementValue: ptrInt(1)},
		{Name: "Bronze III", RequirementType: &levelType, RequirementValue: ptrInt(5)},
		{Name: "Bronze II", RequirementType: &levelType, RequirementValue: ptrInt(9)},
		{Name: "Listener", RequirementType: &ratingType, RequirementValue: ptrInt(10)},
		{Name: "Melody Seeker", RequirementType: &ratingType, RequirementValue: ptrInt(50)},
		{Name: "Special", RequirementType: nil}, // Should always stay
	}

	filtered := domain.FilterHighestBadges(badges)

	assert.Len(t, filtered, 3)

	names := make(map[string]bool)
	for _, b := range filtered {
		names[b.Name] = true
	}

	assert.True(t, names["Bronze II"], "Should keep highest level badge")
	assert.True(t, names["Melody Seeker"], "Should keep highest rating badge")
	assert.True(t, names["Special"], "Should keep special badge")
	assert.False(t, names["Bronze IV"], "Should discard lower level badge")
}

func ptrInt(i int) *int {
	return &i
}
