package task

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type TaskStatus string

var (
	StatusInQueue    TaskStatus = "queue"
	StatusProcessing TaskStatus = "processing"
	StatusDone       TaskStatus = "done"
	StatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID       string     `json:"id"`
	Created  time.Time  `json:"created"`
	Status   TaskStatus `json:"status"`
	Type     string     `json:"type"`
	Target   string     `json:"target"`
	Scanners []string   `json:"scanners"`
}

// TODO use github.com/openshift/oscanner/pkg/error for errors

func (t Task) WriteTo(fd io.Writer) error {
	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("Unable to marshal task: %v", err)
	}

	written, err := fd.Write(data)
	if err != nil || len(data) != written {
		return fmt.Errorf("Unable to write data file: %v", err)
	}

	return nil
}

func (t Task) WriteToFile(filename string) error {
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	return t.WriteTo(fd)
}

func (t *Task) ReadFrom(fd io.Reader) error {
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return fmt.Errorf("Unable to read task data: %v", err)
	}

	if err := json.Unmarshal(data, t); err != nil {
		return fmt.Errorf("Unable to decode task: %v", err)
	}

	return nil
}

func (t *Task) ReadFromFile(filename string) error {
	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	return t.ReadFrom(fd)
}
