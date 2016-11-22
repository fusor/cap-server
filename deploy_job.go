package main

import (
	"fmt"
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

	counter := 0
	for counter != 20 {
		time.Sleep(time.Duration(time.Millisecond * 1000))
		d.emit(fmt.Sprintf("Ticker %d\n", counter))
		counter++
	}

	d.emit(fmt.Sprintf("finished.\n"))
}

func (d *DeployJob) emit(msg string) {
	d.msgBuffer <- fmt.Sprintf("[%s] %s/%s %s",
		d.jobToken, d.registry, d.nuleculeId, msg,
	)
}
