package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v1"

	//"github.com/codeskyblue/go-sh"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const MAIN_FILE = "Nulecule"

// TODO: create a struct
type Answers map[string]map[string]string

func runCommand(cmd string, args ...string) []byte {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Println("Error running " + cmd)
	}
	return output
}

// returns a map of maps
func parseBasicINI(data string) map[string]map[string]string {
	/*
		find first [ then find matching ]. Everything between them is the first key. Read until next [ or end of string.
	*/
	var answers = make(map[string]map[string]string)
	values := strings.SplitAfter(data, "\n")
	var key string
	for _, str := range values {
		if strings.HasPrefix(str, "[") {
			key = strings.Trim(str, "[]\n")
			answers[key] = make(map[string]string)
		} else {
			subvalue := strings.Split(str, " = ")
			answers[key][subvalue[0]] = strings.Trim(subvalue[1], "\n")
		}
	}

	fmt.Println(answers)
	return answers
}

// Nulecule structs
type Param struct {
	Name        string
	Description string
	Default     string
	Binds       []string
}

type Node struct {
	Name   string
	Source string
	Params []Param
}

type Nulecule struct {
	Specversion string
	Id          string
	Metadata    map[string]string
	Params      map[string]string
	Graph       []Node
}

type Bindings struct {
	Src     string `json:"src"`
	SrcKey  string `json:"src_key"`
	Dest    string `json:"dest"`
	DestKey string `json:"dest_key"`
}

// End Nulecule structs

type NuleculeDetail struct {
	Nulecule Answers    `json:"nulecule"`
	Bindings []Bindings `json:"bindings"`
}

func getBindings(nulecule_path string) []Bindings {
	//func getBindings(nulecule_path string) {
	nulecule_file := "nulecule-library/" + nulecule_path + "/Nulecule"
	nulecule, err := ioutil.ReadFile(nulecule_file)
	if err != nil {
		log.Fatal(err)
	}
	n := Nulecule{}
	err = yaml.Unmarshal(nulecule, &n)
	if err != nil {
		log.Fatal(err)
	}
	bindings := make([]Bindings, 0)
	for _, graph := range n.Graph {
		for _, param := range graph.Params {
			for _, bind := range param.Binds {
				bindval := strings.Split(bind, "::")
				b := Bindings{graph.Name, param.Name, bindval[0], bindval[1]}
				bindings = append(bindings, b)
			}
		}
	}
	return bindings
}

func getAnswersFromFile(nulecule_path string) Answers {
	os.Remove("answers.conf")
	/*
		output, err := exec.Command("atomicapp", "genanswers", "nulecule-library/"+nulecule_path).CombinedOutput()
		if err != nil {
			fmt.Println("Error running atomicapp")
		}
	*/
	output := runCommand("atomicapp", "genanswers", "nulecule-library/"+nulecule_path)
	fmt.Println(string(output))
	answers, err := ioutil.ReadFile("answers.conf")
	if err != nil {
		log.Fatal(err)
	}
	return parseBasicINI(string(answers))
}

func getNuleculeList() map[string][]string {
	files, _ := ioutil.ReadDir("./nulecule-library")
	nulecules := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			nulecules = append(nulecules, f.Name())
		}
	}
	return map[string][]string{"nulecules": nulecules}
}

func Nulecules(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered Nulecules method")
	json.NewEncoder(w).Encode(getNuleculeList())
}

func NuleculeDetails(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered NuleculeDetails method")
	vars := mux.Vars(r)
	nulecule_id := vars["id"]
	details := NuleculeDetail{getAnswersFromFile(nulecule_id), getBindings(nulecule_id)}
	json.NewEncoder(os.Stdout).Encode(details)
	json.NewEncoder(w).Encode(details)
}

func genUUID() string {
	return strings.Trim(string(runCommand("/usr/bin/uuidgen")), "\n")
}

func getToken() string {
	return strings.Trim(string(runCommand("/usr/bin/oc", "whoami", "-t")), "\n")
}

func createNewProject(project string) string {
	return strings.Trim(string(runCommand("/usr/bin/oc", "new-project", project)), "\n")
}

func addProviderDetails(answers Answers) {
	uuid := genUUID()
	token := getToken()
	project_name := "cap-" + uuid
	output := createNewProject(project_name)
	fmt.Println(output)
	provider := make(map[string]string)
	provider["namespace"] = project_name
	provider["provider"] = "openshift"
	provider["provider-api"] = "https://10.1.2.2:8443"
	provider["provider-auth"] = token
	provider["provider-cafile"] = "/host/var/lib/openshift/openshift.local.config/master/ca.crt"
	provider["providertlsverify"] = "False"
	answers["general"] = provider
}

func NuleculeUpdate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered NuleculeUpdate method")
	// update the nulecule answers file
	vars := mux.Vars(r)
	nulecule_id := vars["id"]
	fmt.Println(nulecule_id)
	fmt.Println("NuleculeUpdate!")

	// get the posted answers
	// Answers is a map of maps
	res_map := make(map[string]Answers)

	json.NewDecoder(r.Body).Decode(&res_map)

	// ERIK TODO:
	// -> Convert answer JSON params -> map[string]interface{}
	// -> answerMap := addProviderDetails(map[string]interface{}) < adds provider necessary details to [general]
	// -> iniStruct := genINIFromAnswers(answerMap)
	// -> iniStruct.write(/* target nulecule directory */
	addProviderDetails(res_map["nulecule"])

	home_dir := getHomeDir()
	answers_dir := path.Join(home_dir, "answers", nulecule_id)
	os.MkdirAll(answers_dir, 0755)

	f, err := os.Create(path.Join(answers_dir, "answers.conf"))
	if err != nil {
		fmt.Println("Error creating answers.conf")
	}

	defer f.Close()

	for k, v := range res_map["nulecule"] {
		//fmt.Print("[" + k + "]\n")
		fmt.Fprint(f, "["+k+"]\n")
		for k1, v1 := range v {
			//fmt.Printf("%s=%s\n", k1, v1)
			fmt.Fprintf(f, "%s=%s\n", k1, v1)
		}
	}

	json.NewEncoder(w).Encode(res_map) // Success, fail?
}

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func NuleculeDeploy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered NuleculeDeploy method")
	vars := mux.Vars(r)
	nulecule_id := vars["id"]

	home_dir := getHomeDir()

	// Create nulecules dir if it doens't already exist
	nulecules_dir := path.Join(home_dir, "nulecules")
	mode := os.FileMode(int(0755))
	os.Mkdir(nulecules_dir, mode)

	nulecule_dir := path.Join(nulecules_dir, nulecule_id)

	// Download atomicapp
	download_script := path.Join(mainGoDir(), "download_atomicapp.sh")
	output := runCommand("bash", download_script, nulecule_id)
	fmt.Println(string(output))

	// Fix the fact that the entire thing is owned by root -.- WHY
	output = runCommand(
		"sudo", "chown", "-R", "vagrant:vagrant", nulecule_dir)
	fmt.Println(string(output))

	// Copy in generated answers.conf from $HOME/answers working directory
	answers_conf_src := path.Join(home_dir, "answers", nulecule_id, "answers.conf")
	output = runCommand("cp", answers_conf_src, nulecule_dir)
	fmt.Println(string(output))

	// Run the atomicapp!
	run_script := path.Join(mainGoDir(), "run_atomicapp.sh")
	output = runCommand("bash", run_script, nulecule_id)
	fmt.Println(string(output))

	// TODO: EXPOSE ROUTE!
	// Need to figure out a way to tie the "svc" that was just
	// created with the atomicapp that was deployed so we can
	// expose the route correctly.
	//
	// `oc get svc`
	// `oc expose service etherpad-svc -l name=etherpad`

	// TODO: Error handling!
	res_map := make(map[string]interface{})
	res_map["result"] = "success"

	json.NewEncoder(w).Encode(res_map) // Success, fail?
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/index.html")
	t.Execute(w, nil)
}

func wrapScriptCmd(cmd string) string {
	return fmt.Sprintf("\"%s\"", cmd)
}

func mainGoDir() string {
	/*
		_, filename, _, _ := runtime.Caller(0)
		return fmt.Sprintf(path.Dir(filename))
	*/
	return "."
}

func main() {
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/nulecules", Nulecules)
	r.HandleFunc("/nulecules/{id}", NuleculeDetails).Methods("GET")
	r.HandleFunc("/nulecules/{id}", NuleculeUpdate).Methods("POST")
	r.HandleFunc("/nulecules/{id}/deploy", NuleculeDeploy).Methods("POST")

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
