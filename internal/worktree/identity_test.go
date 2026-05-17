package worktree

import (
	"path/filepath"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"my-feature":      "my-feature",
		"feature/Foo Bar": "feature-foo-bar",
		"  ROA-1234  ":    "roa-1234",
		"weird@@chars!!":  "weird-chars",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDirNameSlugRoundTrip(t *testing.T) {
	const pattern = "demo.{slug}"
	for _, slug := range []string{"foo", "roa-1234-bar", "x"} {
		dir := DirName(slug, pattern)
		got, ok := SlugFromDir(dir, pattern)
		if !ok {
			t.Fatalf("SlugFromDir(%q) failed to match pattern", dir)
		}
		if got != slug {
			t.Errorf("round trip: slug %q -> dir %q -> slug %q", slug, dir, got)
		}
	}
}

func TestSlugFromDirRejectsNonMatching(t *testing.T) {
	if _, ok := SlugFromDir("other-thing", "demo.{slug}"); ok {
		t.Error("non-matching directory should not yield a slug")
	}
	if _, ok := SlugFromDir("demo.", "demo.{slug}"); ok {
		t.Error("empty slug should be rejected")
	}
}

func TestResolveAndIdentityFromPath(t *testing.T) {
	const pattern = "demo.{slug}"
	root := filepath.Join(t.TempDir(), "worktrees")

	id := Resolve("my-feat", root, pattern)
	if id.Slug != "my-feat" || id.Dir != "demo.my-feat" {
		t.Fatalf("Resolve gave %+v", id)
	}
	if id.Path != filepath.Join(root, "demo.my-feat") {
		t.Errorf("Resolve path = %q", id.Path)
	}

	back, ok := IdentityFromPath(id.Path, pattern)
	if !ok {
		t.Fatal("IdentityFromPath failed to match")
	}
	if back.Slug != id.Slug {
		t.Errorf("IdentityFromPath slug = %q, want %q", back.Slug, id.Slug)
	}
}
