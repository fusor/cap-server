package main

import (
	"github.com/ventu-io/go-shortid"
)

type Work interface {
	Run(token string, msgBuffer chan<- WorkMsg)
}

type WorkEngine struct {
	msgBuffer chan WorkMsg
}

func NewWorkEngine(bufferSize int) *WorkEngine {
	return &WorkEngine{
		msgBuffer: make(chan WorkMsg, bufferSize),
	}
}

func (engine *WorkEngine) StartNewJob(work Work) string {
	jobToken, _ := shortid.Generate()
	go work.Run(jobToken, engine.msgBuffer)
	return jobToken
}

func (engine *WorkEngine) AttachSubscriber(subscriber WorkSubscriber) {
	subscriber.Subscribe(engine.msgBuffer)
}
