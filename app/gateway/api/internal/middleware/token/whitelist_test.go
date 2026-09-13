package token

import "testing"

func TestWhitelistContains(t *testing.T) {
	w := NewWhitelist([]string{
		"/users",
		" /sessions ", // 首尾空白应被忽略
		"",            // 空项应被忽略
		"/videos/*",
		"/users/*/videos",
	})

	cases := []struct {
		name string
		path string
		want bool
	}{
		{name: "exact hit", path: "/users", want: true},
		{name: "exact miss", path: "/users/1", want: false},
		{name: "trimmed hit", path: "/sessions", want: true},
		{name: "prefix star hit", path: "/videos/popular", want: true},
		{name: "prefix star requires separator", path: "/videos", want: false},
		{name: "middle star hit", path: "/users/42/videos", want: true},
		{name: "middle star needs a segment", path: "/users/videos", want: false},
		{name: "middle star needs full suffix", path: "/users/42/videos/1", want: false},
		{name: "middle star anchors suffix at tail", path: "/users/videos/x/videos", want: true},
		{name: "case sensitive", path: "/Users", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := w.Contains(tc.path); got != tc.want {
				t.Fatalf("Contains(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestWhitelistNilContains(t *testing.T) {
	var w *Whitelist
	if w.Contains("/users") {
		t.Fatal("nil whitelist must not match anything (fail-closed)")
	}
}

func TestDefaultPublicPaths(t *testing.T) {
	w := NewWhitelist(DefaultPublicPaths)

	public := []string{
		"/users",
		"/sessions",
		"/sessions/refresh",
		"/users/42/videos",
		"/videos/popular",
		"/videos/search",
	}
	for _, path := range public {
		if !w.Contains(path) {
			t.Errorf("default whitelist should contain %q", path)
		}
	}

	protected := []string{
		"/videos/publish",
		"/users/42",
	}
	for _, path := range protected {
		if w.Contains(path) {
			t.Errorf("default whitelist should not contain %q", path)
		}
	}
}
