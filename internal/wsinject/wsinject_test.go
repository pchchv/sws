package wsinject

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pchchv/sws/helpers/ancli"
)

const mockHtml = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title></title>
    <link href="css/style.css" rel="stylesheet">
  </head>
  <body>

  </body>
</html>`

type testFileSystem struct {
	root             string
	rootDirFilePaths []string
	nestedDir        string
}

func (tfs *testFileSystem) addRootFile(t *testing.T, suffix string) string {
	t.Helper()
	fileName := fmt.Sprintf("file_%v%v", len(tfs.rootDirFilePaths), suffix)
	path := path.Join(tfs.root, fileName)
	if err := os.WriteFile(path, []byte(mockHtml), 0o777); err != nil {
		t.Fatalf("failed to write root file: %v", err)
	}

	tfs.rootDirFilePaths = append(tfs.rootDirFilePaths, path)
	return path
}

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

func Test_Setup(t *testing.T) {
	tmpDir := t.TempDir()
	ancli.Newline = true
	fileName := "t0.html"
	os.WriteFile(path.Join(tmpDir, fileName), []byte(mockHtml), 0o777)
	nestedDir := path.Join(tmpDir, "nested")
	if err := os.MkdirAll(nestedDir, 0o777); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	nestedFile := path.Join(nestedDir, "nested.html")
	os.WriteFile(nestedFile, []byte(mockHtml), 0o777)
	fs := NewFileServer(8080, "/delta-streamer-ws.js", false, false)
	if _, err := fs.Setup(tmpDir); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	checkIfInjected := func(t *testing.T, filePath string) {
		b, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read mirrored file: %v", err)
		}
		if !strings.Contains(string(b), "delta-streamer.js") {
			t.Fatalf("expected mirrored file: '%v' to have been injected with content-streamer.js", filePath)
		}
	}
	t.Run("it should inject delta-streamer.js source tag", func(t *testing.T) {
		mirrorFilePath := path.Join(fs.mirrorPath, fileName)
		checkIfInjected(t, mirrorFilePath)
	})

	t.Run("it should inject delta-streamer.js souce tag to nested files", func(t *testing.T) {
		mirrorFilePath := path.Join(fs.mirrorPath, "nested", "nested.html")
		checkIfInjected(t, mirrorFilePath)
	})

	checkIfDeltaStreamerExists := func(t *testing.T, filePath string) {
		t.Helper()
		b, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to find delta-streamer.js: %v", err)
		}
		// whatever happens in the delta streamre source code, it should mention sws
		if !strings.Contains(string(b), "sws") {
			t.Fatal("expected delta-streamer.js file to conain string 'sws'")
		}
	}
	t.Run("it should write the delta streamer file to root of mirror", func(t *testing.T) {
		mirrorFilePath := path.Join(fs.mirrorPath, "delta-streamer.js")
		checkIfDeltaStreamerExists(t, mirrorFilePath)
	})
}

func Test_Start(t *testing.T) {
	setup := func(t *testing.T) (*Fileserver, testFileSystem) {
		t.Helper()
		tmpDir := t.TempDir()
		nestedDir := path.Join(tmpDir, "nested")
		err := os.MkdirAll(nestedDir, 0o777)
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		return NewFileServer(8080, "/delta-streamer-ws.js", false, false), testFileSystem{
			root:      tmpDir,
			nestedDir: nestedDir,
		}
	}

	t.Run("it should break on context cancel", func(t *testing.T) {
		fs, _ := setup(t)
		_, err := fs.Setup(t.TempDir())
		if err != nil {
			t.Fatalf("failed to setup test fileserver: %v", err)
		}

		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		t.Cleanup(cancel)

		done := make(chan struct{})
		go func() {
			fs.Start(timeoutCtx)
			close(done)
		}()

		select {
		case <-done:
			// if the fs starts, we expect it to break on context cancel.
			t.Fatalf("expected fs to break on context cancel, but it did not")
		case <-timeoutCtx.Done():
			// timeout means the context was cancelled as expected
		}
	})

	t.Run("file changes", func(t *testing.T) {
		setupReadyFs := func(t *testing.T) (testFileSystem, chan error, chan string, context.Context) {
			t.Helper()
			fs, testFileSystem := setup(t)
			testFileSystem.addRootFile(t, "")
			fs.Setup(testFileSystem.root)
			refreshChan := make(chan string)
			fs.registerWs("mock", refreshChan)
			timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			t.Cleanup(cancel)
			earlyFail := make(chan error, 1)
			awaitFsStart := make(chan struct{})
			go func() {
				close(awaitFsStart)
				err := fs.Start(timeoutCtx)
				if err != nil {
					earlyFail <- err
				}
			}()

			<-awaitFsStart
			// Give the Start a moment to actually start, not just the routine
			time.Sleep(time.Millisecond)
			return testFileSystem, earlyFail, refreshChan, timeoutCtx
		}

		t.Run("it should send a reload event on file changes", func(t *testing.T) {
			testFileSystem, earlyFail, refreshChan, timeoutCtx := setupReadyFs(t)
			testFile := testFileSystem.rootDirFilePaths[0]
			os.WriteFile(testFile, []byte("changes!"), 0o755)

			select {
			case err := <-earlyFail:
				t.Fatalf("start failed: %v", err)
			case got := <-refreshChan:
				expected := "/" + filepath.Base(testFile)
				if got != expected {
					t.Fatalf("expected reload event to be %v, but got %v", expected, got)
				}
			case <-timeoutCtx.Done():
				t.Fatal("failed to receive refresh within time")
			}
		})

		t.Run("it should send a reload event on file additions", func(t *testing.T) {
			testFileSystem, earlyFail, refreshChan, timeoutCtx := setupReadyFs(t)
			testFile := testFileSystem.addRootFile(t, "")
			select {
			case err := <-earlyFail:
				t.Fatalf("start failed: %v", err)
			case got := <-refreshChan:
				expected := "/" + filepath.Base(testFile)
				if got != expected {
					t.Fatalf("expected reload event to be %v, but got %v", expected, got)
				}
			case <-timeoutCtx.Done():
				t.Fatal("failed to receive refresh within time")
			}
		})
	})
}
