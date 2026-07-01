package domain

import "testing"

func TestStringArrayValueNilIsEmptyPostgresArray(t *testing.T) {
	var arr StringArray
	val, err := arr.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if val != "{}" {
		t.Fatalf("expected empty Postgres array, got %#v", val)
	}
}

func TestStringArrayValuePreservesEntries(t *testing.T) {
	arr := StringArray{"Bocchi the Rock", "ぼっち・ざ・ろっく！"}
	val, err := arr.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if val != `{"Bocchi the Rock","ぼっち・ざ・ろっく！"}` {
		t.Fatalf("unexpected serialized array: %#v", val)
	}
}
