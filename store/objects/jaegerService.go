package objects

import (
	"fmt"
	"hash/fnv"

	"github.com/jaegertracing/jaeger/model"
)

const serviceLogType = "jaegerService"

//JaegerService type, for query purposes
type JaegerService struct {
	OperationName string `json:"operationName"`
	ServiceName   string `json:"serviceName"`
	Type          string `json:"type"`
}

//NewLogzioService creates a new logzio service from a span
func NewLogzioService(span *model.Span) JaegerService {
	service := JaegerService{
		ServiceName:   span.Process.ServiceName,
		OperationName: span.OperationName,
		Type:          serviceLogType,
	}
	return service
}

// HashCode receives a logzio service and returns a hash representation of it's service name and operation name.
func (service *JaegerService) HashCode() (string, error) {
	hash := fnv.New64a()
	_, err := hash.Write(append([]byte(service.ServiceName), []byte(service.OperationName)...))
	return fmt.Sprintf("%x", hash.Sum64()), err
}
