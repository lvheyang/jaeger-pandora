package store

import (
	"encoding/json"
	"jaeger-pandora/store/objects"
	"strings"
	"time"

	"github.com/jaegertracing/jaeger/pkg/cache"
	pandora "github.com/lvheyang/pandora-go"

	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
)

const (
	dropLogsDiskThreshold = 98
)

type loggerWriter struct {
	logger hclog.Logger
}

//this is to convert between jaeger log messages and pandoraSender log messages
func (writer *loggerWriter) Write(msgBytes []byte) (n int, err error) {
	msgString := string(msgBytes)
	if strings.Contains(strings.ToLower(msgString), "error") {
		writer.logger.Error(msgString)
	} else {
		writer.logger.Debug(msgString)
	}
	return len(msgBytes), nil
}

// PandoraSpanWriter is a struct which holds pandora span writer properties
type PandoraSpanWriter struct {
	accountToken string
	logger       hclog.Logger
	sender       *pandora.PandoraSender
	serviceCache cache.Cache
	sourcetype   string
	repo         string
}

// NewLogzioSpanWriter creates a new pandora span writer for jaeger
func NewLogzioSpanWriter(config PandoraConfig, logger hclog.Logger) (*PandoraSpanWriter, error) {
	sender, err := pandora.New(
		config.AccountToken,
		pandora.SetUrl(config.ListenerURL()),
		pandora.SetDebug(&loggerWriter{logger: logger}),
		pandora.SetDrainDiskThreshold(dropLogsDiskThreshold))

	if err != nil {
		return nil, err
	}
	spanWriter := &PandoraSpanWriter{
		accountToken: config.AccountToken,
		logger:       logger,
		sender:       sender,
		sourcetype:   config.SourceType,
		repo:         config.Repo,
		serviceCache: cache.NewLRUWithOptions(
			100000,
			&cache.Options{
				TTL: 24 * time.Hour,
			},
		),
	}
	return spanWriter, nil
}

// WriteSpan receives a Jaeger span, converts it to pandora span and sends it to pandora
func (spanWriter *PandoraSpanWriter) WriteSpan(span *model.Span) error {
	spanBytes, err := objects.TransformToJaegerSpanBytes(span)
	if err != nil {
		return err
	}
	req := pandora.PandoraReqBody{
		Raw:        string(spanBytes),
		SourceType: spanWriter.sourcetype,
		Repo:       spanWriter.repo,
		Timestamp:  time.Now().UnixNano() / int64(time.Millisecond),
	}
	err = spanWriter.sender.SendData(req)
	if err != nil {
		return err
	}
	service := objects.NewLogzioService(span)
	serviceHash, err := service.HashCode()

	if spanWriter.serviceCache.Get(serviceHash) == nil || err != nil {
		if err == nil {
			spanWriter.serviceCache.Put(serviceHash, serviceHash)
		}
		serviceBytes, err := json.Marshal(service)
		if err != nil {
			return err
		}
		req := pandora.PandoraReqBody{
			Raw:        string(serviceBytes),
			SourceType: spanWriter.sourcetype,
			Repo:       spanWriter.repo,
			Timestamp:  time.Now().UnixNano() / int64(time.Millisecond),
		}
		err = spanWriter.sender.SendData(req)
	}
	return err
}

// Close stops and drains pandora sender
func (spanWriter *PandoraSpanWriter) Close() {
	spanWriter.sender.Stop()
}
