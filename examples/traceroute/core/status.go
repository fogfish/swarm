//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package core

import (
	"os"

	"github.com/fogfish/curie/v2"
	"github.com/fogfish/swarm"
)

func GetStatus(req Request) ActStatus {
	host := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if host == "" {
		host = "unknown"
	}

	return ActStatus{
		Meta: &swarm.Meta{
			Sink: curie.IRI(req.Channel),
		},
		Data: &Status{
			Host:   host,
			Opaque: req.Opaque,
		},
	}
}
