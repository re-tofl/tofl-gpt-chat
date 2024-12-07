package depgraph

import (
	"go.uber.org/zap"
	"io"
	"os"
)

type DepGraph struct {
	logger      *dgEntity[*zap.SugaredLogger]
	fileWriters map[string]*dgEntity[io.WriteCloser]
}

func NewDepGraph() *DepGraph {
	return &DepGraph{
		logger:      &dgEntity[*zap.SugaredLogger]{},
		fileWriters: make(map[string]*dgEntity[io.WriteCloser]),
	}
}

func (d *DepGraph) GetLogger() (*zap.SugaredLogger, error) {
	return d.logger.get(func() (*zap.SugaredLogger, error) {
		logger := zap.Must(zap.NewDevelopment()).Sugar()
		return logger, nil
	})
}

func (d *DepGraph) GetFileWriter(name string) (io.WriteCloser, error) {
	if fw, ok := d.fileWriters[name]; ok {
		return fw.get(func() (io.WriteCloser, error) {
			return os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		})
	}
	return nil, nil
}
