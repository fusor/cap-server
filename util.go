package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v1"
)

const ANSWERS_FILE = "answers.conf"         // file produced by genanswers
const ANSWERS_FILE_GEN = "answers.conf.gen" // Answers file w/ user provided answers

func genUUID() string {
	return strings.Trim(string(runCommand("/usr/bin/uuidgen")), "\n")
}

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func getNuleculesDir() string {
	return path.Join(getHomeDir(), "nulecules")
}

func getNuleculeDir(registry string, nuleculeId string) string {
	return path.Join(getNuleculesDir(), registry, nuleculeId)
}

func getNuleculeList() NuleculeList {
	nuleculeListFile, _ := ioutil.ReadFile("./nulecule_list.yaml")
	nuleculeList := NuleculeList{}
	err := yaml.Unmarshal(nuleculeListFile, &nuleculeList)
	if err != nil {
		log.Fatal(err)
	}
	return nuleculeList
}

func findEntry(answers Answers, entry string) map[string]string {

	var res_map = make(map[string]string)

	for _, v := range answers {
		for k1, v1 := range v {
			if k1 == entry {
				res_map[k1] = v1
			}
		}
	}
	return res_map
}

func writeUserAnswersToFile(
	registry string,
	nuleculeId string,
	res_map map[string]Answers,
) {
	nuleculeDir := getNuleculeDir(registry, nuleculeId)
	answersFile := path.Join(nuleculeDir, ANSWERS_FILE_GEN)

	f, err := os.Create(answersFile)
	if err != nil {
		log.Fatal("Error writing user answers")
	}

	defer f.Close()

	// Actually write dict out to file in ini format
	for k, v := range res_map["nulecule"] {
		//fmt.Print("[" + k + "]\n")
		fmt.Fprint(f, "["+k+"]\n")
		for k1, v1 := range v {
			//fmt.Printf("%s=%s\n", k1, v1)
			fmt.Fprintf(f, "%s=%s\n", k1, v1)
		}
	}
}

func getGeneratedAnswersFile(registry string, nuleculeId string) Answers {
	nuleculeDir := getNuleculeDir(registry, nuleculeId)
	answersFile := path.Join(nuleculeDir, ANSWERS_FILE_GEN)
	answers, err := ioutil.ReadFile(answersFile)
	if err != nil {
		log.Fatal(err)
	}
	return parseBasicINI(string(answers))
}

// Command helpers
func runCommand(cmd string, args ...string) []byte {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Println("Error running " + cmd)
	}
	return output
}

// Parsing helpers
func getAnswersFromFile(registry string, nuleculeId string) Answers {
	nuleculeDir := getNuleculeDir(registry, nuleculeId)
	answersFile := path.Join(nuleculeDir, ANSWERS_FILE)
	answers, err := ioutil.ReadFile(answersFile)
	if err != nil {
		log.Fatal(err)
	}
	return parseBasicINI(string(answers))
}

type StrippedNamespace struct {
	namespace string
	nodeName  string
}

func stripNamespaces(answers Answers) ([]StrippedNamespace, Answers) {
	var namespace, nodeName string
	var strippedNamespaces = []StrippedNamespace{}

	for answerKey, _ := range answers {
		re, _ := regexp.Compile("^(.*):(.*)$")
		matchGroups := re.FindStringSubmatch(answerKey)

		//matchGroups[0] is the full string, matchGroups[...] are the extracted vals
		if len(matchGroups) != 3 {
			continue
		}

		// Replace namespaced key with stripped key
		namespace = matchGroups[1]
		nodeName = matchGroups[2]
		answer := answers[answerKey]
		delete(answers, answerKey)
		answers[namespace] = answer

		strippedNamespaces = append(strippedNamespaces,
			StrippedNamespace{namespace, nodeName})
	}

	//return namespace, nodeName, answers
	return strippedNamespaces, answers
}

func injectNamespaces(
	namespaceManifest NamespaceManifest,
	answers Answers,
	registry string,
	nuleculeId string,
) {
	atomicAppId := NewAtomicAppId(registry, nuleculeId)
	strippedNamespaces, contains := namespaceManifest[atomicAppId]

	// If no namespaces were stripped, nothing needs to be done
	if !contains {
		return
	}

	// Iterate over stripped namespaces and add them back into answer sections
	for strippedNamespace, strippedNodeName := range strippedNamespaces {
		for section, sectionAnswers := range answers {
			// If answers contains section header that matches a stripped namespace,
			// create fully qualified section header and add it back to the answers
			if section != strippedNamespace {
				continue
			}

			fqSection := fmt.Sprintf("%s:%s", strippedNamespace, strippedNodeName)
			delete(answers, section)
			answers[fqSection] = sectionAnswers
		}
	}
}

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
			subvalue := strings.Split(str, "=")
			if len(subvalue) > 1 {
				answers[key][strings.TrimSpace(subvalue[0])] = strings.TrimSpace(subvalue[1])
			}
		}
	}

	//fmt.Println(answers)
	return answers
}

func getBindings(registry string, nuleculeId string) []Bindings {
	nuleculeFile := path.Join(getNuleculeDir(registry, nuleculeId), "Nulecule")
	nulecule, err := ioutil.ReadFile(nuleculeFile)
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

func addProviderDetails(answers Answers) string {
	uuid := genUUID()
	token := getToken()
	projectName := "cap-" + uuid
	provider := make(map[string]string)
	provider["namespace"] = projectName
	provider["provider"] = "openshift"
	provider["provider-api"] = "https://10.1.2.2:8443"
	provider["provider-auth"] = token
	provider["provider-cafile"] = "/host/var/lib/openshift/openshift.local.config/master/ca.crt"
	provider["providertlsverify"] = "False"
	answers["general"] = provider
	return projectName
}

// Openshift helpers
func getToken() string {
	return strings.Trim(string(runCommand("/usr/bin/oc", "whoami", "-t")), "\n")
}

func createNewProject(project string) string {
	return strings.Trim(string(runCommand("/usr/bin/oc", "new-project", project)), "\n")
}

// Atomic helpers
func downloadNulecule(registry string, nuleculeId string) {
	download_script := path.Join(".", "download_atomicapp.sh")
	output := runCommand("bash", download_script, registry, nuleculeId)
	fmt.Println(string(output))
}

func pingHost(host string) int {
	fmt.Println(host)

	u, err := url.Parse(host)
	if err != nil {
		fmt.Println("could not parse host, should probably return a 400 BadRequest")
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	// ping the host
	resp, err := http.Get(u.String())
	if err != nil {
		// TODO: do we print out the errors? or just ignore them
		return 0
	}

	return resp.StatusCode
}
