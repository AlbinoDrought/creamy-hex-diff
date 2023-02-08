package main

import (
	"io"
	"os"
)

type creamyFile struct {
	path           string
	source         *os.File
	info           os.FileInfo
	offset         int64
	readDataLength int
	buffer         []byte
}

func (f *creamyFile) IsAtBeginning() bool {
	return f.offset <= 0
}

func (f *creamyFile) lastBufferPos() int64 {
	pos := f.info.Size() - int64(len(f.buffer))
	return pos + (0x10 - (pos % int64(0x10))) // ceil to 0x10 offset
}

func (f *creamyFile) IsAtEnd() bool {
	return f.offset >= f.lastBufferPos()
}

func (f *creamyFile) At(offset int64) (int, error) {
	f.offset = offset
	return f.Read()
}

func (f *creamyFile) Read() (int, error) {
	f.source.Seek(f.offset, io.SeekStart)
	length, err := f.source.Read(f.buffer)
	f.readDataLength = length
	return length, err
}

func (f *creamyFile) Start() (int, error) {
	f.offset = 0
	return f.Read()
}

func (f *creamyFile) End() (int, error) {
	f.offset = f.lastBufferPos()
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
	if f.IsAtEnd() {
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
