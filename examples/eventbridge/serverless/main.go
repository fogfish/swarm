package main

import (
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/queue/eventbridge"
)

func main() {
	eventbridge.NewServerlessApp("swarm-example-eventbridge").
		CreateEventBus().
		CreateSink(
			&eventbridge.SinkProps{
				Queue: "swarm-test",
				Lambda: &scud.FunctionGoProps{
					SourceCodePackage: "github.com/fogfish/swarm",
					SourceCodeLambda:  "examples/eventbridge/recv",
				},
			},
		).
		Synth(nil)
}
