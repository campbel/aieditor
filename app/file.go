package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/quick"
	"github.com/fsnotify/fsnotify"

	"github.com/campbel/aieditor/log"
)

type FileBuffer struct {
	mu sync.RWMutex

	path    string
	content string
	display string
}

func NewFile(path string) *FileBuffer {
	f := &FileBuffer{path: path}
	f.load()
	return f
}

func (f *FileBuffer) Path() string {
	return f.path
}

func (f *FileBuffer) Content() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.content
}

func (f *FileBuffer) Display() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.display
}

func (f *FileBuffer) Set(content string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	content = strings.TrimSpace(strings.Replace(content, "\t", "    ", -1))
	f.content = content
	var b bytes.Buffer
	err := quick.Highlight(&b, f.content, GetLanguage(f.path), "terminal16m", "dracula")
	if err != nil {
		f.display = f.content
	}
	f.display = b.String()
}

func (f *FileBuffer) Save() error {
	return os.WriteFile(f.path, []byte(f.Content()), 0644)
}

func (f *FileBuffer) Watch(fn func()) error {

	dir := filepath.Dir(f.path)
	log.Debug("watching", "dir", dir)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				log.Debug("watch event", "event", event, "ok", ok)
				if !ok {
					return
				}
				if event.Name == f.path && event.Has(fsnotify.Write) {
					f.load()
					fn()
				}
			case err, ok := <-watcher.Errors:
				log.Debug("watch error", "error", err, "ok", ok)
			}
		}
	}()

	return watcher.Add(dir)
}

func (f *FileBuffer) load() error {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return err
	}
	f.Set(string(data))
	return nil
}
