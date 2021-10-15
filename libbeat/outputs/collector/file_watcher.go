package collector

import (
	"fmt"
	"strings"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/fsnotify/fsnotify"
)

const (
	clusterCredentialPath   = "/erda-cluster-credential"
	clusterCredentialKey    = "CLUSTER_ACCESS_KEY"
	newCredentialCreateFlag = "data"
)

var (
	clusterCredentialFullPath = fmt.Sprintf("%s/%s", clusterCredentialPath, clusterCredentialKey)
)

type FileWatcher struct {
	log      *logp.Logger
	filePath string
	watcher  *fsnotify.Watcher
	onEvent  func(event fsnotify.Event)
}

func NewFileWatcher(log *logp.Logger, onEvent func(event fsnotify.Event)) (*FileWatcher, error) {
	var err error

	f := &FileWatcher{
		log:      log,
		filePath: clusterCredentialPath,
		onEvent:  onEvent,
	}
	f.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err = f.watch(); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *FileWatcher) Close() error {
	return f.watcher.Close()
}

func (f *FileWatcher) watch() error {
	f.log.Infof("starting watch credential, path: %s", clusterCredentialPath)
	go func(filePath string) {
		for {
			select {
			case event, ok := <-f.watcher.Events:
				if !ok {
					return
				}

				f.log.Infof("credential file event: %v", event)

				switch event.Op {
				// kubernetes secret mount content change will invoke: CREATE->CHMOD->RENAME->CREATE->REMOVE:
				// watch event CREATE which last time and path contain `data`, first CREATE event will be backup old credential.
				case fsnotify.Create:
					// filter
					if strings.Contains(event.Name, newCredentialCreateFlag) {
						f.onEvent(event)
					}
				default:
					continue
				}

			case err, ok := <-f.watcher.Errors:
				if !ok {
					return
				}
				f.log.Errorf("file watcher error: %v", err)
				return
			}
		}
	}(f.filePath)

	return f.watcher.Add(f.filePath)
}
