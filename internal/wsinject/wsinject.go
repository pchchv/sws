package wsinject

import (
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

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
