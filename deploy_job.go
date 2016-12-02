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
	host       string
	msgBuffer  chan<- IWorkMsg
	jobToken   string
}

func NewDeployJob(registry string, nuleculeId string, host string) *DeployJob {
	return &DeployJob{
		registry:   registry,
		nuleculeId: nuleculeId,
		host:       host,
	}
}

func (d *DeployJob) Run(jobToken string, msgBuffer chan<- IWorkMsg) {
	d.jobToken = jobToken
	d.msgBuffer = msgBuffer

	d.emit("Executing deployment script...\n")
	d.runDeploymentScript()

	d.emit("Initiating health check...")
	d.runHealthCheck(300 * time.Second)

	d.emit("Full deployment finished.")
}

func (d *DeployJob) runDeploymentScript() {
	runScript := path.Join(".", "run_atomicapp.sh")
	output := runCommand("bash", runScript, d.registry, d.nuleculeId)
	fmt.Println(string(output))

	d.emit("Deployment script executed successfully!")
}

func (d *DeployJob) runHealthCheck(timeout time.Duration) {
	var statuscode int
	start := time.Now()
	var elapsed time.Duration

	for statuscode != 200 || elapsed > timeout {
		statuscode := pingHost(d.host)
		if statuscode == 200 {
			d.emit("Pinged the service, 200! It's up!")
		} else {
			d.emit("Pinged the service")
		}
		// sleep 1s
		time.Sleep(1 * time.Second)
		elapsed = time.Since(start)
	}

	d.emit("Health Check timed out!")
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
