package database

import (
	"fmt"
	"os"

	"github.com/openshift/oscanner/pkg/configuration"
)

type DirType int

const (
	TaskNew DirType = iota
	TaskQueue
	TaskProcessing
	TaskDone
	TaskFailed
	TaskCache
)

var (
	subdirs = map[DirType]string{
		TaskNew:        "new",
		TaskQueue:      "queue",
		TaskProcessing: "processing",
		TaskDone:       "done",
		TaskFailed:     "failed",
		TaskCache:      "cache",
	}
)

type DB struct {
	config *configuration.Configuration
}

func NewDB(config *configuration.Configuration) *DB {
	return &DB{
		config: config,
	}
}

func (db *DB) basedir(suffix string) string {
	return fmt.Sprintf("%s/%s", db.config.Storage.Path, suffix)
}

func (db *DB) dir(suffix, id string) string {
	return fmt.Sprintf("%s/%s/%s", db.config.Storage.Path, suffix, id)
}

func (db *DB) BaseDir(dirType DirType) string {
	dir, ok := subdirs[dirType]
	if !ok {
		panic(fmt.Sprintf("unknown type: %v", dirType))
	}
	return db.basedir(dir)
}

func (db *DB) TaskDir(dirType DirType, id string) string {
	dir, ok := subdirs[dirType]
	if !ok {
		panic(fmt.Sprintf("unknown type: %v", dirType))
	}
	return db.dir(dir, id)
}

func (db *DB) TaskData(dirType DirType, id string) string {
	return fmt.Sprintf("%s/data", db.TaskDir(dirType, id))
}

func (db *DB) Init() error {
	for k := range subdirs {
		dir := db.BaseDir(k)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Unable to make directory: %s: %v", dir, err)
		}
	}
	return nil
}
