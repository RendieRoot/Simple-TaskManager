package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// JSONController - json controller
func JSONController(taskID string, data string, command string) string {
	jsonFile := "tasks.json"

	type TaskStructure struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Assignee    string `json:"assignee"`
		Date        string `json:"date"`
		Status      string `json:"status"`
	}

	tasks := make([]TaskStructure, 0)
	tasksJSON, err := ioutil.ReadFile(jsonFile)

	if err != nil {
		log.Print("[ERROR] There was a problem reading the json file")
	}

	json.Unmarshal(tasksJSON, &tasks)

	switch command {
	case "return_all":
		output, _ := json.Marshal(tasks)
		log.Print("[INFO] Returned all tasks")
		return string(output)
	case "return_spec":
		for i := range tasks {
			if taskID == tasks[i].ID {
				output, _ := json.Marshal(tasks[i])
				return string(output)
				log.Print("[INFO] Returned task, ID: ", taskID)
			}
		}
		return "400 - ID not found"
	case "delete":
		tmpTasks := make([]TaskStructure, 0)
		for i := range tasks {
			if taskID != tasks[i].ID {
				tmpTasks = append(tmpTasks, tasks[i])
			}
		}

		if len(tmpTasks) != len(tasks) {
			jsonToFile, _ := json.Marshal(tmpTasks)
			ioutil.WriteFile(jsonFile, jsonToFile, os.ModePerm)
			log.Print("[INFO] Task removed")
			return "Done"
		}
		return "400 - ID not found"
	case "create":
		tmpTask := TaskStructure{}
		json.Unmarshal([]byte(data), &tmpTask)
		tasks = append(tasks, tmpTask)

		jsonToFile, _ := json.Marshal(tasks)
		ioutil.WriteFile(jsonFile, jsonToFile, os.ModePerm)
		log.Print("[INFO] Task created, ID", taskID)

		return data
	case "change":
		tmpTasks := make([]TaskStructure, 0)

		for i := range tasks {
			if taskID != tasks[i].ID {
				tmpTasks = append(tmpTasks, tasks[i])
			} else {
				tmpTask := TaskStructure{}
				json.Unmarshal([]byte(data), &tmpTask)
				tmpTasks = append(tmpTasks, tmpTask)
				log.Print("[INFO] Task changed, ID: ", taskID)
			}
		}

		jsonToFile, _ := json.Marshal(tmpTasks)
		ioutil.WriteFile(jsonFile, jsonToFile, os.ModePerm)
		return data
	}
	log.Print("[INFO] Unknown command: ", command)
	return "400 - Unknown command"
}

// GETRequest -  returns all tasks in JSON format (use “encoding/json“ package) OR returns only one task with specified ID (if such task does not exist, return 404)
func GETRequest(id string) string {
	if id == "" {
		return JSONController("", "", "return_all")
	}

	return JSONController(id, "", "return_spec")
}

// POSTRequest - create new task (you should supply a task in request body in JSON format)
func POSTRequest(data string) string {
	return JSONController("", data, "create")
}

// PATCHRequest - replace the task with specified ID with the task supplied in the request body
func PATCHRequest(id, data string) string {
	return JSONController(id, data, "change")
}

// DELETERequest - removed task with specified ID from the list
func DELETERequest(id string) string {
	return JSONController(id, "", "delete")
}

// RequestHandler - handle requests
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	id := mux.Vars(r)["id"]
	data, _ := ioutil.ReadAll(r.Body)

	log.Print("[INFO] Handled request with method: ", method, ", with id: ", id)

	switch method {
	case "GET":
		fmt.Fprint(w, GETRequest(id))
	case "POST":
		fmt.Fprint(w, POSTRequest(string(data)))
	case "PATCH":
		fmt.Fprint(w, PATCHRequest(id, string(data)))
	case "DELETE":
		fmt.Fprint(w, DELETERequest(id))
	default:
		log.Print("[WARNING] Unsuitable method requested: ", method)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "400 - Unsuitable method requested")
	}
}

func main() {
	socket := ":8585"

	r := mux.NewRouter()
	r.HandleFunc("/api/tasks", RequestHandler)
	r.HandleFunc("/api/tasks/{id}", RequestHandler)
	http.Handle("/", r)

	log.Print("[INFO] Server is listening... Socket: ", socket)

	http.ListenAndServe(socket, nil)
}
