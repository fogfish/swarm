package main

import (
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/queue/eventsqs"
)

func main() {
	eventsqs.NewServerlessApp("swarm-example-sqs").
		CreateQueue().
		CreateSink(
			&eventsqs.SinkProps{
				Lambda: &scud.FunctionGoProps{
					SourceCodePackage: "github.com/fogfish/swarm",
					SourceCodeLambda:  "examples/sqs/recv",
				},
			},
		).
		Synth(nil)
}
