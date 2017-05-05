package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/distribution/uuid"

	"github.com/openshift/oscanner/pkg/database"
	targetfactory "github.com/openshift/oscanner/pkg/image/factory"
	"github.com/openshift/oscanner/pkg/instance"
	"github.com/openshift/oscanner/pkg/server/response"
	"github.com/openshift/oscanner/pkg/task"
)

type TaskRequest struct {
	Type     string   `json:"type"`
	Target   string   `json:"target"`
	Scanners []string `json:"scanners"`
}

func ListTasksHandler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	db := database.NewDB(in.Config)

	ans := []task.Task{}

	for _, dirname := range []database.DirType{
		database.TaskQueue,
		database.TaskProcessing,
		database.TaskDone,
		database.TaskFailed,
	} {
		err := filepath.Walk(db.BaseDir(dirname), func(path string, info os.FileInfo, err error) error {
			if info.IsDir() || !strings.HasSuffix(path, "/data") {
				return nil
			}

			t := task.Task{}

			if err := t.ReadFromFile(path); err != nil {
				return fmt.Errorf("Unable to read task: %v", err)
			}

			ans = append(ans, t)

			return nil
		})
		if err != nil {
			response.HTTPResponse(w, 500, "%v", err)
			return
		}
	}

	data, err := json.Marshal(ans)
	if err != nil {
		response.HTTPResponse(w, 500, "Unable to marshal tasks: %v", err)
		return
	}

	w.Write([]byte(`{"result":`))
	w.Write(data)
	w.Write([]byte(`}`))
}

func CreateTaskHandler(in *instance.Instance, w http.ResponseWriter, r *http.Request) {
	taskRequest := &TaskRequest{}

	data, err := ioutil.ReadAll(io.LimitReader(r.Body, in.Config.Tasks.MaxSize))
	if err != nil {
		response.HTTPResponse(w, 400, "Unable to read task: %v", err)
		return
	}

	if err := json.Unmarshal(data, taskRequest); err != nil {
		response.HTTPResponse(w, 400, "Unable to decode task: %v", err)
		return
	}

	if len(taskRequest.Type) > 0 {
		receivers := targetfactory.List()
		validType := false
		for _, v := range receivers {
			if v == taskRequest.Type {
				validType = true
				break
			}
		}
		if !validType {
			response.HTTPResponse(w, 400, "Unknown task type '%s' (expected %v)", taskRequest.Type, receivers)
			return
		}
	} else {
		response.HTTPResponse(w, 400, "Empty task type")
		return
	}

	if len(taskRequest.Target) == 0 {
		response.HTTPResponse(w, 400, "Empty task target")
		return
	}

	id := uuid.Generate().String()

	db := database.NewDB(in.Config)
	newTask := db.TaskDir(database.TaskNew, id)

	if err := os.MkdirAll(newTask, 0755); err != nil {
		response.HTTPResponse(w, 500, "Unable to make directory: %v", err)
		return
	}

	t := task.Task{
		ID:      id,
		Created: time.Now(),
		Status:  task.StatusInQueue,
		Type:    taskRequest.Type,
		Target:  taskRequest.Target,
	}

	err = t.WriteToFile(db.TaskData(database.TaskNew, id))
	if err != nil {
		os.RemoveAll(newTask)
		response.HTTPResponse(w, 500, "Unable to write task data: %v", err)
		return
	}

	if err := os.Rename(newTask, db.TaskDir(database.TaskQueue, id)); err != nil {
		os.RemoveAll(newTask)
		response.HTTPResponse(w, 500, "Unable to move task to processing: %v", err)
		return
	}

	in.Dispatcher.Awake()

	w.Write([]byte(`{"result":"` + id + `"}`))
}
