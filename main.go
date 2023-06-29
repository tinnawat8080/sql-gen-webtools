package main

import (
	"bufio"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var tmpl = template.Must(template.ParseGlob("templates/*"))
var sqlFilesPath = "/Users/5311637/workspace/backend/initial-sql"

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")

	http.HandleFunc("/", showListHandler)
	http.HandleFunc("/db-list/search", getDbListHandler)
	http.HandleFunc("/sql-generate", generateSQLHandler)
	http.ListenAndServe(":"+port, nil)

}

func showListHandler(w http.ResponseWriter, r *http.Request) {
	showList := ShowList{}
	entries, err := os.ReadDir(sqlFilesPath)
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
		dbDir, _ := os.ReadDir(sqlFilesPath + "/" + sprint + "/" + template)
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
	var allContent string
	for _, sprint := range generateSQLRequest.Sprints {
		var sprintContent string = ""
		for _, db := range generateSQLRequest.DBs {
			var path = sqlFilesPath + "/" + sprint + "/" + template + "/" + db
			files, _ := os.ReadDir(path)
			for _, file := range files {
				body, _ := os.ReadFile(path + "/" + file.Name())
				sprintContent += string(body)
			}
		}
		value, _ := os.Open(sqlFilesPath + "/" + sprint + "/" + "value_local.txt")
		fileScanner := bufio.NewScanner(value)
		fileScanner.Split(bufio.ScanLines)
		var valueMap map[string]string = make(map[string]string)
		for fileScanner.Scan() {
			v := strings.Split(fileScanner.Text(), "=")
			valueMap[v[0]] = v[1]
		}
		sprintContent = replaceVariableWithValue(sprintContent, valueMap)
		allContent += sprintContent
	}
	response, _ := json.Marshal(allContent)
	w.Write(response)
}

func replaceVariableWithValue(content string, valueMap map[string]string) string {
	for key, element := range valueMap {
		content = strings.ReplaceAll(content, "$"+key+"$", element)
	}
	return content
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
