package fsdump

import (
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
)

type File struct {
	io.Reader
	io.Closer
	Path string
}

type FileSource interface {
	GetFile() *File
}

type ChannelFileSource struct {
	Chan chan File
}

func (source *ChannelFileSource) GetFile() *File {
	file, ok := <-source.Chan
	if !ok {
		return nil
	}
	return &file
}

type Dumper struct {
	Src        FileSource
	Dest       string
	NumWorkers int
}

func (dumper *Dumper) Run() {
	if dumper.Src == nil {
		panic("FileSource cannot be nil")
	}

	numWorkers := dumper.NumWorkers
	if numWorkers < 1 {
		numWorkers = runtime.NumCPU()
	}

	wg := sync.WaitGroup{}
	wg.Add(numWorkers)
	for range numWorkers {
		worker := worker{
			dumper: dumper,
			buffer: make([]byte, 4096),
		}
		go func() {
			worker.run()
			wg.Done()
		}()
	}
	wg.Wait()
}

type worker struct {
	dumper *Dumper
	buffer []byte
}

func (worker *worker) run() {
	for file := worker.dumper.Src.GetFile(); file != nil; file = worker.dumper.Src.GetFile() {
		if err := worker.dumpFile(file); err != nil {
			log.Printf("failed to dump file: %s: %v", file.Path, err)
		}
		file.Close()
	}
}

func (worker *worker) dumpFile(in *File) error {
	dir, _ := path.Split(in.Path)
	if err := os.MkdirAll(path.Join(worker.dumper.Dest, dir), 0700); err != nil {
		return err
	}

	out, err := os.Create(path.Join(worker.dumper.Dest, in.Path))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.CopyBuffer(out, in, worker.buffer)
	return err
}
