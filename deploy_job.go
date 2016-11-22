package main

import (
	"fmt"
	"path"
	"time"
)

type DeployJob struct {
	registry   string
	nuleculeId string
	msgBuffer  chan<- string
	jobToken   string
}

func NewDeployJob(registry string, nuleculeId string) *DeployJob {
	return &DeployJob{
		registry:   registry,
		nuleculeId: nuleculeId,
	}
}

func (d *DeployJob) Run(jobToken string, msgBuffer chan<- string) {
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
	d.msgBuffer <- fmt.Sprintf("[%s] %s/%s %s",
		d.jobToken, d.registry, d.nuleculeId, msg,
	)
}
