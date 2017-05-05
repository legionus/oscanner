package dispatcher

import (
	"io"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/openshift/oscanner/pkg/configuration"
	"github.com/openshift/oscanner/pkg/database"
	targetfactory "github.com/openshift/oscanner/pkg/image/factory"
	scanfactory "github.com/openshift/oscanner/pkg/scanner/factory"
	"github.com/openshift/oscanner/pkg/task"
	"github.com/openshift/oscanner/pkg/types"

	_ "github.com/openshift/oscanner/pkg/scanner/filecontant"

	_ "github.com/openshift/oscanner/pkg/image/dockerimage"
)

type Dispatcher struct {
	config *configuration.Configuration
	awake  chan struct{}
}

func NewDispatcher(c *configuration.Configuration) *Dispatcher {
	return &Dispatcher{
		config: c,
		awake:  make(chan struct{}),
	}
}

func (d *Dispatcher) Awake() {
	select {
	case d.awake <- struct{}{}:
	default:
	}
}

func (d *Dispatcher) Run() error {
	db := database.NewDB(d.config)

	dir, err := os.Open(db.BaseDir(database.TaskQueue))
	if err != nil {
		return err
	}
	defer dir.Close()

	for {
		_, err = dir.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		taskIDs, err := dir.Readdirnames(-1)
		if err != nil {
			return err
		}

		for _, taskID := range taskIDs {
			// TODO limit parallel goroutines
			go d.process(taskID)
		}

		<-d.awake
	}
}

func (d *Dispatcher) process(taskID string) {
	log.Debugf("processing task %s", taskID)

	db := database.NewDB(d.config)

	queueTask := db.TaskDir(database.TaskQueue, taskID)
	processingTask := db.TaskDir(database.TaskProcessing, taskID)

	t := &task.Task{}

	if err := t.ReadFromFile(queueTask); err != nil {
		log.Errorf("unable to read task %s: %v", taskID, err)
		return
	}

	moveToFailed := func() {
		t.Status = task.StatusFailed

		if err := t.WriteToFile(processingTask); err != nil {
			log.Errorf("unable to update failed task  %s: %v", taskID, err)
		}

		if err := os.Rename(queueTask, db.TaskDir(database.TaskFailed, taskID)); err != nil {
			log.Errorf("unable to move task %s to failed queue: %v", taskID, err)
		}
	}

	moveToDone := func() {
		log.Infof("task %s is done", taskID)

		t.Status = task.StatusDone

		if err := t.WriteToFile(processingTask); err != nil {
			log.Errorf("unable to update done task %s: %v", taskID, err)
		}

		if err := os.Rename(queueTask, db.TaskDir(database.TaskDone, taskID)); err != nil {
			log.Errorf("unable to move task %s to archive: %v", taskID, err)
		}
	}

	cacheDir := db.TaskDir(database.TaskCache, taskID)

	receiver, err := targetfactory.Get(t.Type, d.config, t.Target, cacheDir)
	if err != nil {
		log.Errorf("unable to get target receiver for task %s: %v", taskID, err)
		moveToFailed()
		return
	}

	scanners := []types.Scanner{}

	for _, v := range t.Scanners {
		scanner, err := scanfactory.Get(v, d.config)
		if err != nil {
			log.Errorf("unable to get scanner for task %s: %v", taskID, err)
			moveToFailed()
			return
		}
		scanners = append(scanners, scanner)
	}

	if len(scanners) == 0 {
		moveToDone()
		return
	}

	t.Status = task.StatusProcessing

	if err := t.WriteToFile(processingTask); err != nil {
		log.Errorf("unable to update task %s: %v", taskID, err)
		return
	}

	if err := os.Rename(queueTask, processingTask); err != nil {
		log.Errorf("unable to move task %s to processing queue: %v", taskID, err)
		return
	}

	if err := receiver.Fetch(); err != nil {
		log.Errorf("unable to fetch target for task %s: %v", taskID, err)
		moveToFailed()
		return
	}

	if err := receiver.Snapshot(); err != nil {
		log.Errorf("unable to make snapshot for task %s: %v", taskID, err)
		moveToFailed()
		return
	}

	for i, scanner := range scanners {
		if err := scanner.Run(cacheDir); err != nil {
			log.Errorf("scanner %s failed: %v", t.Scanners[i], err)
			moveToFailed()
		}
	}

	receiver.Cleanup()
	moveToDone()
}
