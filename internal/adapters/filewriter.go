package adapters

import (
	"fmt"
	"os"
	"sync"
)

type FileWriter struct {
	mu   sync.RWMutex
	file *os.File
}

func InitFileWriter(name string) (*FileWriter, error) {
	file, err := os.OpenFile(name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла: %w", err)
	}
	return &FileWriter{
		file: file,
	}, nil
}

func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.file.Close()
}

func (fw *FileWriter) Write(data []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.file.Write(data)
}

func (fw *FileWriter) TruncateAndWrite(data []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.file.WriteAt(data, 0)
}

func (fw *FileWriter) Read() ([]byte, error) {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return os.ReadFile(fw.file.Name())
}
