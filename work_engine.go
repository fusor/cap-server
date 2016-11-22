package main

import (
	"github.com/ventu-io/go-shortid"
)

type IWork interface {
	Run(token string, msgBuffer chan<- IWorkMsg)
}

type WorkEngine struct {
	msgBuffer chan IWorkMsg
}

func NewWorkEngine(bufferSize int) *WorkEngine {
	return &WorkEngine{
		msgBuffer: make(chan IWorkMsg, bufferSize),
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
