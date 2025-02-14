package wsinject

import (
	"os"
	"path"
	"slices"
	"testing"
)

func Test_walkDir(t *testing.T) {
	t.Run("it should visit every file ", func(t *testing.T) {
		var got []string
		want := []string{"t0", "t1", "t2"}
		tmpDir := t.TempDir()
		os.WriteFile(path.Join(tmpDir, "t0"), []byte("t0"), 0o644)
		os.WriteFile(path.Join(tmpDir, "t1"), []byte("t1"), 0o644)
		nestedDir, err := os.MkdirTemp(tmpDir, "dir0_*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		os.WriteFile(path.Join(nestedDir, "t2"), []byte("t2"), 0o644)

		wsInjectMaster(tmpDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				t.Fatalf("got err during traversal: %v", err)
			}
			if d.IsDir() {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			got = append(got, string(b))
			return nil
		})

		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})
}
