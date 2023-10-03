package store

import (
	"errors"
	"net/url"
)

var factories = make([]SourceFactory, 0)

var ErrUnsupportedSourceScheme = errors.New("unsupported source name")

type SourceFactory func(uri *url.URL) (Source, error)

func RegisterSourceFactory(factory SourceFactory) {
	factories = append(factories, factory)
}

func CreateSource(uri *url.URL) (Source, error) {
	for _, factory := range factories {
		source, err := factory(uri)
		if err != nil {
			return nil, err
		}
		if errors.Is(err, ErrUnsupportedSourceScheme) {
			continue
		}

		return source, nil
	}

	return nil, ErrUnsupportedSourceScheme
}
