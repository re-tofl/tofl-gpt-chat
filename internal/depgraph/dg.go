package depgraph

import (
	"github.com/re-tofl/tofl-gpt-chat/internal/adapters"
	"go.uber.org/zap"
)

type DepGraph struct {
	logger      *dgEntity[*zap.SugaredLogger]
	fileWriters map[string]*dgEntity[*adapters.FileWriter]
}

func NewDepGraph() *DepGraph {
	return &DepGraph{
		logger:      &dgEntity[*zap.SugaredLogger]{},
		fileWriters: make(map[string]*dgEntity[*adapters.FileWriter]),
	}
}

func (d *DepGraph) GetLogger() (*zap.SugaredLogger, error) {
	return d.logger.get(func() (*zap.SugaredLogger, error) {
		logger := zap.Must(zap.NewDevelopment()).Sugar()
		return logger, nil
	})
}

func (d *DepGraph) GetFileWriter(name string) (*adapters.FileWriter, error) {
	fw, ok := d.fileWriters[name]
	if !ok {
		fw = &dgEntity[*adapters.FileWriter]{}
		d.fileWriters[name] = fw
	}
	return fw.get(func() (*adapters.FileWriter, error) {
		return adapters.InitFileWriter(name)
	})
}
