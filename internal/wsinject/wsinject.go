package wsinject

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var ErrNoHeaderTagFound = errors.New("no header tag found")

type Fileserver struct {
	masterPath            string
	mirrorPath            string
	forceReload           bool
	expectTLS             bool
	wsPort                int
	wsPath                string
	watcher               *fsnotify.Watcher
	pageReloadChan        chan string
	wsDispatcher          sync.Map
	wsDispatcherStarted   *bool
	wsDispatcherStartedMu *sync.Mutex
}

func NewFileServer(wsPort int, wsPath string, forceReload, expectTLS bool) *Fileserver {
	mirrorDir, err := os.MkdirTemp("", "sws_*")
	if err != nil {
		panic(err)
	}

	started := false
	return &Fileserver{
		mirrorPath:            mirrorDir,
		wsPort:                wsPort,
		wsPath:                wsPath,
		expectTLS:             expectTLS,
		forceReload:           forceReload,
		pageReloadChan:        make(chan string),
		wsDispatcher:          sync.Map{},
		wsDispatcherStarted:   &started,
		wsDispatcherStartedMu: &sync.Mutex{},
	}
}

func wsInjectMaster(root string, do func(path string, d fs.DirEntry, err error) error) error {
	if err := filepath.WalkDir(root, do); err != nil {
		log.Fatalf("Error walking the path %q: %v\n", root, err)
	}
	return nil
}

func injectScript(html []byte, scriptTag string) ([]byte, error) {
	htmlStr := string(html)
	// find the location of the closing `</header>` tag
	idx := strings.Index(htmlStr, "</head>")
	if idx == -1 {
		return html, ErrNoHeaderTagFound
	}

	var buf bytes.Buffer
	// write the HTML up to the closing `</head>` tag
	if _, err := buf.WriteString(htmlStr[:idx]); err != nil {
		return nil, fmt.Errorf("failed to write pre: %w", err)
	}

	if _, err := buf.WriteString(scriptTag); err != nil {
		return nil, fmt.Errorf("failed to write script tag: %w", err)
	}

	if _, err := buf.WriteString(htmlStr[idx:]); err != nil {
		return nil, fmt.Errorf("failed to write post: %w", err)
	}

	return buf.Bytes(), nil
}
