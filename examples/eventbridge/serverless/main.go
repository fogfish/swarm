//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"github.com/fogfish/scud"
	"github.com/fogfish/swarm/queue/eventbridge"
)

func main() {
	app := eventbridge.NewServerlessApp()

	stack := app.NewStack("swarm-example-eventbridge")
	stack.NewEventBus()

	stack.NewSink(
		&eventbridge.SinkProps{
			Queue: "swarm-test",
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/eventbridge/recv",
			},
		},
	)

	app.Synth(nil)
}
