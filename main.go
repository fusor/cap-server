package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
  "strings"
)

// We need to strip namespaces off answer file sections to talk to the
// front end, but atmoicapp 0.6.4 expects them to be in the answerfile when
// we go to run it, so it will need to be written out to answer.conf.gen before
// running a user's answers. We'll keep track of that bookkeeping here
// TODO: Consider longer term approach than a gross global manifest!
var namespaceManifest NamespaceManifest

func main() {
	namespaceManifest = make(NamespaceManifest)

	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/nulecules", Nulecules).Methods("POST")
	r.HandleFunc("/nulecules/{registry}/{id}", NuleculeDetails).Methods("GET")
	r.HandleFunc("/nulecules/{registry}/{id}", NuleculeUpdate).Methods("POST")
	r.HandleFunc("/nulecules/{registry}/{id}/deploy", NuleculeDeploy).Methods("POST")
	r.HandleFunc("/health-check", RunHealthCheck).Methods("POST")

	// Setup static file server at /static/, used for stuff like js
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
	r.PathPrefix("/static/").Handler(fs)

	// Serve index template
	r.HandleFunc("/", IndexHandler)

	fmt.Println("Listening on localhost:3001")
	allowed_headers := handlers.AllowedHeaders([]string{"Content-Type"})
	log.Fatal(http.ListenAndServe(":3001", handlers.CORS(
		allowed_headers,
	)(r)))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/index.html")
	t.Execute(w, nil)
}

func Nulecules(w http.ResponseWriter, r *http.Request) {
  fmt.Println("Entered Nulecules method")

  res_map := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&res_map)

  organization := res_map["org"]
  username := res_map["username"]
  password := res_map["password"]

  channel := make(chan string)
  
  list := getNuleculeList(organization, username, password)
  fmt.Println("Length:", len(list.Nulecules))
  responses := make([]string, len(list.Nulecules))

  filtered_list := NuleculeList{}
  for _, nules := range list.Nulecules {
    go IsImageAtomicApp("docker://" + nules, channel)
  }

  counter := 1
  for response := range channel {
    responses = append(responses, response)
    if counter != len(list.Nulecules) {
      counter++
    } else {
      fmt.Println("Closing Channel")
      close(channel)
    }
  }

  for _, value := range responses {
    if strings.Compare(value,"ignore\n") != 0 && value != "" {
      filtered_list.Nulecules = append(filtered_list.Nulecules, value)
    }
  }
  fmt.Println(filtered_list.Nulecules)
  json.NewEncoder(w).Encode(filtered_list)
}

func NuleculeDetails(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered NuleculeDetails method")
	vars := mux.Vars(r)
	registry := vars["registry"]
	nuleculeId := vars["id"]

	downloadNulecule(registry, nuleculeId)

	// atomicapp 0.6.4 changed the answer.conf file format and namespaces
	// component names with their container name, i.e.
	// "mariadb-app" vs "mariadb-centos7-atomicapp:mariadb-app"
	// stripContainerNamespace will strip off the container namespace
	// to maintain backwards compatibility for the cap UI.

	strippedNamespaces, answers := stripNamespaces(
		getAnswersFromFile(registry, nuleculeId),
	)

	for _, strippedNamespace := range strippedNamespaces {
		namespaceManifest.insert(registry, nuleculeId,
			strippedNamespace.namespace, strippedNamespace.nodeName)
	}

	details := NuleculeDetail{
		answers,
		getBindings(registry, nuleculeId),
	}

	json.NewEncoder(os.Stdout).Encode(details)
	json.NewEncoder(w).Encode(details)
}

func NuleculeUpdate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered NuleculeUpdate method")
	// update the nulecule answers file
	vars := mux.Vars(r)
	nuleculeId := vars["id"]
	registry := vars["registry"]

	// get the posted answers
	// Answers is a map of maps
	res_map := make(map[string]Answers)
	json.NewDecoder(r.Body).Decode(&res_map)

	// TODO: Consider better way to uniquely ID projects instead of a UUID
	// Could also use UUIDs as bookkeeping on the backend with a more friendly
	// project name provided by the user on the front end.
	projectName := addProviderDetails(res_map["nulecule"])
	createNewProject(projectName)
	injectNamespaces(namespaceManifest, res_map["nulecule"], registry, nuleculeId)
	writeUserAnswersToFile(registry, nuleculeId, res_map)

	json.NewEncoder(w).Encode(res_map) // Success, fail?
}

func NuleculeDeploy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered NuleculeDeploy method")
	vars := mux.Vars(r)
	nulecule_id := vars["id"]
	registry := vars["registry"]

	// Run the atomicapp!
	run_script := path.Join(mainGoDir(), "run_atomicapp.sh")
	output := runCommand("bash", run_script, registry, nulecule_id)
	fmt.Println(string(output))

	// TODO: Error handling!
	res_map := make(map[string]interface{})
	res_map["result"] = "success"

	json.NewEncoder(w).Encode(res_map) // Success, fail?
}

func RunHealthCheck(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&body)
	host := body["host"]

	health_check_script := path.Join(mainGoDir(), "health_check.sh")
	output := runCommand("bash", health_check_script, host)

	outputstr := string(output)

	isAlive := false
	if outputstr == "200" {
		isAlive = true
	}

	res_map := make(map[string]interface{})
	res_map["is_alive"] = isAlive
	json.NewEncoder(w).Encode(res_map)
}
