package main

import (
	"io"
	"os"
)

type creamyFile struct {
	path   string
	source *os.File
	info   os.FileInfo
	offset int64
	buffer []byte
}

func (f *creamyFile) Read() (int, error) {
	f.source.Seek(f.offset, io.SeekStart)
	return f.source.Read(f.buffer)
}

func (f *creamyFile) Start() (int, error) {
	f.offset = 0
	return f.Read()
}

func (f *creamyFile) End() (int, error) {
	f.offset = f.info.Size() - int64(len(f.buffer))
	f.offset = f.offset - (f.offset % int64(len(f.buffer)))
	return f.Read()
}

func (f *creamyFile) Last(bytes int64) (int, error) {
	f.offset -= bytes
	if f.offset < 0 {
		return f.Start()
	}
	return f.Read()
}

func (f *creamyFile) Next(bytes int64) (int, error) {
	f.offset += bytes
	if f.offset+int64(len(f.buffer)) > f.info.Size() {
		return f.End()
	}
	return f.Read()
}

func openCreamyFile(path string, bufferSize int) (*creamyFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &creamyFile{
		path:   path,
		source: file,
		info:   info,
		offset: 0,
		buffer: make([]byte, bufferSize),
	}, nil
}
