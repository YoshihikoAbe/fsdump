package fsdump_test

import (
	"os"
	"path"
	"testing"

	"github.com/YoshihikoAbe/fsdump"
)

const dataRoot = "./testdata/"

var files []os.DirEntry

func init() {
	var err error
	if files, err = os.ReadDir(dataRoot); err != nil {
		panic(err)
	}
}

func TestDump(t *testing.T) {
	temp := t.TempDir()

	dumpFiles(temp)
	for _, file := range files {
		name := file.Name()
		data, err := os.ReadFile(path.Join(temp, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != name+" test" {
			t.Fatal("horrible things have happened")
		}
	}
}

func BenchmarkDump(b *testing.B) {
	temp := b.TempDir()

	for b.Loop() {
		dumpFiles(temp)
	}
}

func dumpFiles(dest string) {
	dumper := &fsdump.Dumper{
		Src:  getFiles(),
		Dest: dest,
	}
	dumper.Run()
}

func getFiles() *fsdump.ChannelFileSource {
	ch := make(chan fsdump.File, 1)

	go func() {
		for _, entry := range files {
			f, err := os.Open(dataRoot + entry.Name())
			if err != nil {
				panic(err)
			}

			ch <- fsdump.File{
				Reader: f,
				Closer: f,
				Path:   entry.Name(),
			}
		}
		close(ch)
	}()

	return &fsdump.ChannelFileSource{
		Chan: ch,
	}
}
