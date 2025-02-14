package wsinject

import (
	"sync"

	"github.com/pchchv/sws/helpers/ancli"
)

func (fs *Fileserver) wsDispatcherStart() {
	for {
		pageToReload, ok := <-fs.pageReloadChan
		if !ok {
			ancli.PrintNotice("stopping wsDispatcher")
			fs.wsDispatcher.Range(func(key, value any) bool {
				ancli.PrintfNotice("sending to: '%v'", key)
				wsWriterChan := value.(chan string)
				// close chan to stop the wsRoutine
				close(wsWriterChan)
				return true
			})
			return
		}
		ancli.PrintfNotice("got update: '%v'", pageToReload)
		fs.wsDispatcher.Range(func(key, value any) bool {
			ancli.PrintfNotice("sending to: '%v'", key)
			wsWriterChan := value.(chan string)
			wsWriterChan <- pageToReload
			return true
		})
	}
}

func (fs *Fileserver) registerWs(name string, c chan string) {
	if !read(fs.wsDispatcherStartedMu, fs.wsDispatcherStarted) {
		go fs.wsDispatcherStart()
		write(fs.wsDispatcherStartedMu, true, fs.wsDispatcherStarted)
	}
	ancli.PrintfNotice("registering: '%v'", name)
	fs.wsDispatcher.Store(name, c)
}

func (fs *Fileserver) deregisterWs(name string) {
	fs.wsDispatcher.Delete(name)
}

// Reads by locking the mutex before taking a copy, will then return the copy.
func read[T any](m *sync.Mutex, src *T) T {
	m.Lock()
	defer m.Unlock()
	return *src
}

// Writes by locking the mutex before writing.
func write[T any](m *sync.Mutex, value T, dest *T) {
	m.Lock()
	defer m.Unlock()
	*dest = value
}
