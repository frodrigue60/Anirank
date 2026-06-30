package postgres

import "testing"

func TestMakeSlug(t *testing.T) {
	if got := makeSlug("PARCO"); got != "parco" {
		t.Errorf("expected parco, got %s", got)
	}
	if got := makeSlug("  Foo Bar  "); got != "foo-bar" {
		t.Errorf("expected foo-bar, got %s", got)
	}
}

func TestIsPQUniqueViolation(t *testing.T) {
	if isPQUniqueViolation(nil, "animes_anilist_id_unique") {
		t.Error("nil error should not match")
	}
}
