package jsondir

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ekklesion/multitenancy/store"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const scheme = "json+dir"

var ErrNotJson = errors.New("not a json file")
var ErrNotDir = errors.New("not a directory")

func init() {
	store.RegisterSourceFactory(CreateJsonSource)
}

// CreateJsonSource creates a jsondir Source
// TODO: Make the filenames unaware of domain names
func CreateJsonSource(uri *url.URL) (store.Source, error) {
	if uri.Scheme != scheme {
		return nil, store.ErrUnsupportedSourceScheme
	}

	path := uri.Path
	if uri.Opaque != "" {
		path = uri.Opaque
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, ErrNotDir
	}

	source := &Source{
		path: path,
	}

	w := uri.Query().Get("watch")
	if w == "true" || w == "1" {
		source.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
		source.done = make(chan struct{})
	}

	return source, nil
}

type Source struct {
	path    string                   // The path to the json directory
	watcher *fsnotify.Watcher        // The notify watcher
	done    chan struct{}            // This channel is for closing the goroutine that watches
	sites   map[string]*store.Tenant // The sites currently loaded
	mtx     sync.Mutex               // Mutex for preventing race conditions writes
}

func (s *Source) GetSite(_ context.Context, r *http.Request) (*store.Tenant, error) {
	if s.sites == nil {
		return nil, store.ErrNotInitialized
	}

	site, ok := s.sites[r.Host]
	if !ok {
		return nil, store.ErrTenantNotFound
	}

	return site, nil
}

func (s *Source) Initialize(_ context.Context) error {
	if s.sites != nil {
		return store.ErrAlreadyInitialized
	}

	err := s.loadInitialSites()
	if err != nil {
		return err
	}

	err = s.watchForFileChanges()
	if err != nil {
		return err
	}

	return nil
}

func (s *Source) Close() error {
	if s.done != nil {
		s.done <- struct{}{}
		close(s.done)
	}

	if s.watcher == nil {
		return nil
	}

	return s.watcher.Close()
}

func (s *Source) loadInitialSites() error {
	s.sites = make(map[string]*store.Tenant)

	files, err := os.ReadDir(s.path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		site, host, err := s.parseSiteInfo(name)
		if err != nil {
			continue
		}

		s.addSite(host, site)
	}

	return nil
}

func (s *Source) parseSiteInfo(name string) (*store.Tenant, string, error) {
	if !strings.HasSuffix(name, ".json") {
		return nil, "", ErrNotJson
	}

	name = strings.TrimPrefix(name, s.path+string(filepath.Separator))
	host := strings.TrimSuffix(name, ".json")

	f, err := os.Open(filepath.Join(s.path, name))
	if err != nil {
		return nil, host, err
	}

	defer func(c io.Closer) {
		_ = c.Close()
	}(f)

	site := &store.Tenant{}
	err = json.NewDecoder(f).Decode(site)
	if err != nil {
		return nil, host, err
	}

	return site, host, nil
}

func (s *Source) addSite(host string, site *store.Tenant) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.sites[host] = site
}

func (s *Source) removeSite(host string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.sites, host)
}

func (s *Source) watchForFileChanges() error {
	if s.watcher == nil {
		return nil // No watching was configured
	}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-s.watcher.Events:
				if !ok {
					continue
				}
				err := s.processEvent(event)
				if err != nil {
					continue
				}

			case _, ok := <-s.done:
				if !ok {
					continue
				}

				return
			}
		}
	}()

	return s.watcher.Add(s.path)
}

func (s *Source) processEvent(evt fsnotify.Event) error {
	switch true {
	// A site has been updated
	case evt.Has(fsnotify.Write):
		site, host, err := s.parseSiteInfo(evt.Name)
		if err != nil {
			return nil
		}

		log.Println("updating site", host)
		s.addSite(host, site)
	// A site has been created
	case evt.Has(fsnotify.Create):
		site, host, err := s.parseSiteInfo(evt.Name)
		if err != nil {
			return nil
		}

		log.Println("adding site", host)
		s.addSite(host, site)

	// A site has been removed
	case evt.Has(fsnotify.Remove):
		_, host, _ := s.parseSiteInfo(evt.Name)
		if host == "" {
			return nil
		}

		log.Println("removing site", host)
		s.removeSite(host)
	// When a file is renamed, we remove it from the entries
	case evt.Has(fsnotify.Rename):
		_, host, _ := s.parseSiteInfo(evt.Name)
		if host == "" {
			return nil
		}

		log.Println("removing site", host)
		s.removeSite(host)
	}

	return nil
}
