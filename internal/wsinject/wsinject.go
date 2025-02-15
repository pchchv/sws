package wsinject

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pchchv/sws/helpers/ancli"
)

const deltaStreamer = `<!-- This script has been injected by sws and allows hot reloads -->
<script type="module" src="delta-streamer.js"></script>`

var ErrNoHeaderTagFound = errors.New("no header tag found")

type Fileserver struct {
	masterPath            string
	mirrorPath            string
	forceReload           bool
	wsPort                int
	wsPath                string
	watcher               *fsnotify.Watcher
	pageReloadChan        chan string
	wsDispatcher          sync.Map
	wsDispatcherStarted   *bool
	wsDispatcherStartedMu *sync.Mutex
}

func NewFileServer(wsPort int, wsPath string, forceReload bool) *Fileserver {
	mirrorDir, err := os.MkdirTemp("", "sws_*")
	if err != nil {
		panic(err)
	}

	started := false
	return &Fileserver{
		mirrorPath:            mirrorDir,
		wsPort:                wsPort,
		wsPath:                wsPath,
		forceReload:           forceReload,
		pageReloadChan:        make(chan string),
		wsDispatcher:          sync.Map{},
		wsDispatcherStarted:   &started,
		wsDispatcherStartedMu: &sync.Mutex{},
	}
}

func (fs *Fileserver) Setup(pathToMaster string) (string, error) {
	ancli.PrintfNotice("mirroring root: '%v'", pathToMaster)
	fs.masterPath = pathToMaster
	watcher, err := fsnotify.NewWatcher()
	fs.watcher = watcher
	if err != nil {
		return "", fmt.Errorf("failed to create fsnotify watcher: %e", err)
	}

	if err = wsInjectMaster(pathToMaster, fs.mirrorMaker); err != nil {
		return "", fmt.Errorf("failed to create websocket injected mirror: %e", err)
	}

	if err = fs.writeDeltaStreamerScript(); err != nil {
		return "", fmt.Errorf("failed to write delta streamer file: %e", err)
	}

	return fs.mirrorPath, nil
}

// Start starts listening to file events,
// update mirror and stream notifications on which files to update.
func (fs *Fileserver) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case fsEv, ok := <-fs.watcher.Events:
			if !ok {
				return errors.New("fsnotify watcher event channel closed")
			}
			fs.handleFileEvent(fsEv)
		case fsErr, ok := <-fs.watcher.Errors:
			if !ok {
				return errors.New("fsnotify watcher error channel closed")
			}
			return fsErr
		}
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
		return nil, fmt.Errorf("failed to write pre: %e", err)
	}

	if _, err := buf.WriteString(scriptTag); err != nil {
		return nil, fmt.Errorf("failed to write script tag: %e", err)
	}

	if _, err := buf.WriteString(htmlStr[idx:]); err != nil {
		return nil, fmt.Errorf("failed to write post: %e", err)
	}

	return buf.Bytes(), nil
}

func injectWebsocketScript(b []byte) (bool, []byte, error) {
	var injected bool
	contentType := http.DetectContentType(b)
	// only act on html files
	if !strings.Contains(contentType, "text/html") {
		return injected, b, nil
	}

	injected = true
	b, err := injectScript(b, deltaStreamer)
	if err != nil {
		if !errors.Is(err, ErrNoHeaderTagFound) {
			return injected, nil, fmt.Errorf("failed to inject script tag: %e", err)
		} else {
			injected = false
		}
	}

	return injected, b, nil
}

func (fs *Fileserver) writeDeltaStreamerScript() error {
	err := os.WriteFile(
		path.Join(fs.mirrorPath, "delta-streamer.js"),
		[]byte(fmt.Sprintf(deltaStreamerSourceCode, fs.wsPort, fs.wsPath, fs.forceReload)),
		0o755)
	if err != nil {
		return fmt.Errorf("failed to write delta-streamer.js: %e", err)
	}

	return nil
}

func (fs *Fileserver) mirrorFile(origPath string) error {
	relativePath := strings.Replace(origPath, fs.masterPath, "", -1)
	fileB, err := os.ReadFile(origPath)
	if err != nil {
		return fmt.Errorf("failed to read file on path: '%v', err: %v", origPath, err)
	}

	injected, injectedBytes, err := injectWebsocketScript(fileB)
	if err != nil {
		return fmt.Errorf("failed to inject websocket script: %e", err)
	} else if injected {
		ancli.PrintfNotice("injected delta-streamer script loading tag in: '%v'", origPath)
	}

	mirroredPath := path.Join(fs.mirrorPath, relativePath)
	relativePathDir := path.Dir(mirroredPath)
	if err = os.MkdirAll(relativePathDir, 0o755); err != nil {
		return fmt.Errorf("failed to create relative dir: '%v', error: %v", relativePathDir, err)
	}

	if err = os.WriteFile(mirroredPath, injectedBytes, 0o755); err != nil {
		return fmt.Errorf("failed to write mirrored file: %e", err)
	}

	return nil
}

func (fs *Fileserver) mirrorMaker(p string, info os.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err = fs.watcher.Add(p); err != nil {
			return fmt.Errorf("failed to add recursive path: %e", err)
		}
		return nil
	}

	return fs.mirrorFile(p)
}

func (fs *Fileserver) notifyPageUpdate(fileName string) {
	// make filename relative idempotently
	fs.pageReloadChan <- strings.Replace(fileName, fs.masterPath, "", -1)
}

func (fs *Fileserver) handleFileEvent(fsEv fsnotify.Event) {
	if fsEv.Has(fsnotify.Write) {
		ancli.PrintfNotice("noticed file write in orig file: '%s'", fsEv.Name)
		fs.mirrorFile(fsEv.Name)
		fs.notifyPageUpdate(fsEv.Name)
	}
}
