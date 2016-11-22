package main

import (
	"github.com/ventu-io/go-shortid"
)

type IWork interface {
	Run(token string, msgBuffer chan<- string)
}

type WorkEngine struct {
	msgBuffer chan string
}

func NewWorkEngine(bufferSize int) *WorkEngine {
	return &WorkEngine{
		msgBuffer: make(chan string, bufferSize),
	}
}

func (engine *WorkEngine) StartNewJob(work IWork) string {
	jobToken, _ := shortid.Generate()
	go work.Run(jobToken, engine.msgBuffer)
	return jobToken
}

func (engine *WorkEngine) AttachSubscriber(subscriber IWorkSubscriber) {
	subscriber.Subscribe(engine.msgBuffer)
}