package worktree

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"my-feature":      "my-feature",
		"feature/Foo Bar": "feature-foo-bar",
		"  ROA-1234  ":    "roa-1234",
		"weird@@chars!!":  "weird-chars",
	}
	for in, want := range cases {
		require.Equal(t, want, Slugify(in), "Slugify(%q)", in)
	}
}

func TestDirNameSlugRoundTrip(t *testing.T) {
	const pattern = "demo.{slug}"
	for _, slug := range []string{"foo", "roa-1234-bar", "x"} {
		dir := dirName(slug, pattern)
		got, ok := SlugFromDir(dir, pattern)
		require.True(t, ok, "SlugFromDir(%q) should match the pattern", dir)
		require.Equal(t, slug, got, "round trip via dir %q", dir)
	}
}

func TestSlugFromDirRejectsNonMatching(t *testing.T) {
	_, ok := SlugFromDir("other-thing", "demo.{slug}")
	require.False(t, ok, "non-matching directory should not yield a slug")

	_, ok = SlugFromDir("demo.", "demo.{slug}")
	require.False(t, ok, "empty slug should be rejected")
}

func TestResolve(t *testing.T) {
	const pattern = "demo.{slug}"
	root := filepath.Join(t.TempDir(), "worktrees")

	id := Resolve("my-feat", root, pattern)
	require.Equal(t, "my-feat", id.Slug)
	require.Equal(t, "demo.my-feat", id.Dir)
	require.Equal(t, filepath.Join(root, "demo.my-feat"), id.Path)
}
