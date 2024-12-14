package depgraph

import (
	"github.com/re-tofl/tofl-gpt-chat/internal/utils"
	"go.uber.org/zap"
)

type DepGraph struct {
	logger      *dgEntity[*zap.SugaredLogger]
	fileWriters map[string]*dgEntity[*utils.FileWriter]
}

func NewDepGraph() *DepGraph {
	return &DepGraph{
		logger:      &dgEntity[*zap.SugaredLogger]{},
		fileWriters: make(map[string]*dgEntity[*utils.FileWriter]),
	}
}

func (d *DepGraph) GetLogger() (*zap.SugaredLogger, error) {
	return d.logger.get(func() (*zap.SugaredLogger, error) {
		logger := zap.Must(zap.NewDevelopment()).Sugar()
		return logger, nil
	})
}

func (d *DepGraph) GetFileWriter(name string) (*utils.FileWriter, error) {
	fw, ok := d.fileWriters[name]
	if !ok {
		fw = &dgEntity[*utils.FileWriter]{}
		d.fileWriters[name] = fw
	}
	return fw.get(func() (*utils.FileWriter, error) {
		return utils.InitFileWriter(name)
	})
}
