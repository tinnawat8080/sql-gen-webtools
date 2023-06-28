package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var tmpl = template.Must(template.ParseGlob("templates/*"))
var baseFilePath = "/Users/5311637/workspace/backend/initial-sql"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/", showListHandler)
	http.HandleFunc("/db-list/search", getDbListHandler)
	http.HandleFunc("/sql-generate", generateSQLHandler)
	http.ListenAndServe(":"+port, nil)

}

func showListHandler(w http.ResponseWriter, r *http.Request) {
	showList := ShowList{}
	entries, err := os.ReadDir(baseFilePath)
	if err != nil {
		log.Fatal(err)
	}

	var sprintNames []string
	for _, e := range entries {
		if e.IsDir() {
			sprintNames = append(sprintNames, e.Name())
		}
	}
	showList.SprintNames = sprintNames
	tmpl.ExecuteTemplate(w, "show-list", showList)
}

func getDbListHandler(w http.ResponseWriter, r *http.Request) {
	var selectedSprints []string
	json.NewDecoder(r.Body).Decode(&selectedSprints)

	var dbNames []string
	var template = "template"
	for _, sprint := range selectedSprints {
		dbDir, _ := os.ReadDir(baseFilePath + "/" + sprint + "/" + template)
		for _, db := range dbDir {
			if !contains(dbNames, db.Name()) {
				dbNames = append(dbNames, db.Name())
			}
		}
	}

	response, _ := json.Marshal(dbNames)
	w.Write(response)
}

func generateSQLHandler(w http.ResponseWriter, r *http.Request) {
	var generateSQLRequest GenerateSQLRequest
	json.NewDecoder(r.Body).Decode(&generateSQLRequest)
	var template = "template"
	if generateSQLRequest.Action == "rollback" {
		template = "rollback"
	}
	var content string
	for _, sprint := range generateSQLRequest.Sprints {
		for _, db := range generateSQLRequest.DBs {
			var path = baseFilePath + "/" + sprint + "/" + template + "/" + db
			files, _ := os.ReadDir(path)
			for _, file := range files {
				body, _ := os.ReadFile(path + "/" + file.Name())
				content += string(body)
			}
		}
	}
	response, _ := json.Marshal(content)
	w.Write(response)
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

type ShowList struct {
	SprintNames []string
}

type GenerateSQLRequest struct {
	Sprints []string
	DBs     []string
	Action  string
}
