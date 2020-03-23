package store

import (
	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

// Store is span store struct for pandora jaeger span storage
type Store struct {
	reader *JaegerSpanReader
	writer *PandoraSpanWriter
}

// NewLogzioStore creates a new pandora span store for jaeger
func NewLogzioStore(config PandoraConfig, logger hclog.Logger) *Store {
	reader := NewLogzioSpanReader(config, logger)
	writer, err := NewLogzioSpanWriter(config, logger)
	if err != nil {
		logger.Error("Failed to create pandora span writer: " + err.Error())
	}
	store := &Store{
		reader: reader,
		writer: writer,
	}
	return store
}

// Close the span store
func (store *Store) Close() {
	store.writer.Close()
}

// SpanReader returns the created pandora span reader
func (store *Store) SpanReader() spanstore.Reader {
	return store.reader
}

// SpanWriter returns the created pandora span writer
func (store *Store) SpanWriter() spanstore.Writer {
	return store.writer
}

// DependencyReader return the created pandora dependency store
func (store *Store) DependencyReader() dependencystore.Reader {
	return store.reader
}
