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
	"github.com/fogfish/swarm/broker/eventddb"
)

func main() {
	app := eventddb.NewServerlessApp()

	stack := app.NewStack("swarm-example-ddbstream")
	stack.NewGlobalTable()

	stack.NewSink(
		&eventddb.SinkProps{
			Lambda: &scud.FunctionGoProps{
				SourceCodePackage: "github.com/fogfish/swarm",
				SourceCodeLambda:  "examples/eventddb/dequeue",
			},
		},
	)

	app.Synth(nil)
}
