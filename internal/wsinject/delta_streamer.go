package wsinject

import "github.com/pchchv/sws/helpers/ancli"

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
