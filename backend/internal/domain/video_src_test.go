package domain

import "testing"

func TestIsStorageVideoSrc(t *testing.T) {
	path := "videos/2025/winter/foo.webm"
	httpURL := "https://youtube.com/embed/abc"
	empty := ""

	cases := []struct {
		name string
		src  *string
		want bool
	}{
		{name: "nil", src: nil, want: false},
		{name: "empty", src: &empty, want: false},
		{name: "storage key", src: &path, want: true},
		{name: "https legacy", src: &httpURL, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsStorageVideoSrc(tc.src); got != tc.want {
				t.Fatalf("IsStorageVideoSrc() = %v, want %v", got, tc.want)
			}
		})
	}
}
