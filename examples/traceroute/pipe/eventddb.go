//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package pipe

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/fogfish/dynamo/v3/service/ddb"
	"github.com/fogfish/swarm/broker/eventddb"
	"github.com/fogfish/swarm/examples/traceroute/core"
)

var (
	dbbEmit *ddb.Storage[core.RequestDB]
	ddbOnce sync.Once
)

func getDdbEmiter() *ddb.Storage[core.RequestDB] {
	ddbOnce.Do(func() {
		dbbEmit = ddb.Must(
			ddb.New[core.RequestDB](
				os.Getenv(eventddb.EnvConfigTargetDynamoDB),
			),
		)
	})
	return dbbEmit
}

func ToDynamoDB(req core.ReqTrace) error {
	db := getDdbEmiter()
	err := db.Put(context.Background(), core.ToRequestDB(req))
	if err != nil {
		slog.Error("failed to put item in AWS DynamoDB", "err", err)
		return err
	}
	return nil
}
