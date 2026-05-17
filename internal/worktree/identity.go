// Package worktree wraps `git worktree` and derives the stable, path-based
// identity that every croft-managed checkout carries (PLAN.md §2.1).
package worktree

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Identity is a worktree's stable, path-derived identity. The slug is computed
// from the directory name and never changes when git HEAD moves — navigating a
// branch stack must not change a worktree's resources or state keys.
type Identity struct {
	Slug string // stable identity, e.g. "my-feature"
	Dir  string // directory name, e.g. "demo.my-feature"
	Path string // absolute path to the worktree checkout
}

// slugUnsafe matches runs of characters not allowed in a slug.
var slugUnsafe = regexp.MustCompile(`[^a-z0-9._-]+`)

// Slugify converts an arbitrary branch name into a filesystem-safe slug.
func Slugify(branch string) string {
	s := strings.ToLower(strings.TrimSpace(branch))
	s = slugUnsafe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-._")
}

// DirName renders a worktree directory name from a slug and a naming pattern.
// The pattern must contain the {slug} placeholder.
func DirName(slug, pattern string) string {
	return strings.Replace(pattern, "{slug}", slug, 1)
}

// SlugFromDir extracts the slug from a worktree directory name given the
// naming pattern. ok is false when dir does not match the pattern.
func SlugFromDir(dir, pattern string) (slug string, ok bool) {
	pre, post, found := strings.Cut(pattern, "{slug}")
	if !found || len(dir) < len(pre)+len(post) {
		return "", false
	}
	if !strings.HasPrefix(dir, pre) || !strings.HasSuffix(dir, post) {
		return "", false
	}
	slug = dir[len(pre) : len(dir)-len(post)]
	if slug == "" {
		return "", false
	}
	return slug, true
}

// Resolve builds an Identity for a slug given the (absolute) worktree root and
// the naming pattern.
func Resolve(slug, worktreeRoot, pattern string) Identity {
	dir := DirName(slug, pattern)
	return Identity{
		Slug: slug,
		Dir:  dir,
		Path: filepath.Join(worktreeRoot, dir),
	}
}

// IdentityFromPath derives an Identity from an existing worktree path. ok is
// false when the path's directory name does not match the naming pattern.
func IdentityFromPath(path, pattern string) (id Identity, ok bool) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Identity{}, false
	}
	slug, ok := SlugFromDir(filepath.Base(abs), pattern)
	if !ok {
		return Identity{}, false
	}
	return Identity{Slug: slug, Dir: filepath.Base(abs), Path: abs}, true
}
