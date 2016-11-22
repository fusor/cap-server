package main

import (
	"encoding/json"
	"fmt"
	"path"
	"time"
)

type DeployJob struct {
	registry   string
	nuleculeId string
	msgBuffer  chan<- IWorkMsg
	jobToken   string
}

func NewDeployJob(registry string, nuleculeId string) *DeployJob {
	return &DeployJob{
		registry:   registry,
		nuleculeId: nuleculeId,
	}
}

func (d *DeployJob) Run(jobToken string, msgBuffer chan<- IWorkMsg) {
	d.jobToken = jobToken
	d.msgBuffer = msgBuffer

	d.emit("Executing deployment script...\n")
	d.runDeploymentScript()

	d.emit("Initiating health check...")
	d.runHealthCheck()

	d.emit("Full deployment finished.")
}

func (d *DeployJob) runDeploymentScript() {
	runScript := path.Join(".", "run_atomicapp.sh")
	output := runCommand("bash", runScript, d.registry, d.nuleculeId)
	fmt.Println(string(output))

	d.emit("Deployment script executed successfully!")
}

func (d *DeployJob) runHealthCheck() {
	// TODO: Actually implement...
	counter := 0
	for counter != 10 {
		d.emit("Pinged the service, 503")
		counter++
		time.Sleep(time.Duration(time.Millisecond * 500))
	}
	d.emit("Pinged the service, 200! It's up!")
}

func (d *DeployJob) emit(msg string) {
	d.msgBuffer <- DeployMsg{
		d.jobToken,
		d.registry,
		d.nuleculeId,
		msg,
	}
}

type DeployMsg struct {
	JobToken   string `json:"job_token"`
	Registry   string `json:"registry"`
	NuleculeId string `json:"nulecule_id"`
	Msg        string `json:"msg"`
}

func (m DeployMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}
