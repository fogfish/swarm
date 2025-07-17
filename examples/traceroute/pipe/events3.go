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
	"path/filepath"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/curie/v2"
	"github.com/fogfish/dynamo/v3/service/s3"
	"github.com/fogfish/swarm/broker/events3"
	"github.com/fogfish/swarm/examples/traceroute/core"
)

var (
	s3Emit *s3.Storage[core.RequestDB]
	s3Once sync.Once
)

func getS3Emiter(bucket string) *s3.Storage[core.RequestDB] {
	s3Once.Do(func() {
		s3Emit = s3.Must(
			s3.New[core.RequestDB](bucket),
		)
	})
	return s3Emit
}

func ToS3(req core.ReqTrace) error {
	db := getS3Emiter(os.Getenv(events3.EnvConfigTargetS3))
	err := db.Put(context.Background(), core.ToRequestDB(req))
	if err != nil {
		slog.Error("failed to put item in AWS S3", "err", err)
		return err
	}
	return nil
}

func FromS3(evt *events.S3EventRecord) (core.RequestDB, error) {
	db := getS3Emiter(os.Getenv(events3.EnvConfigSourceS3))
	rdb, err := db.Get(context.Background(),
		core.RequestDB{
			HKey: curie.IRI(filepath.Dir(evt.S3.Object.Key)),
			SKey: curie.IRI(filepath.Base(evt.S3.Object.Key)),
		},
	)
	if err != nil {
		slog.Error("failed to read s3 object", "err", err)
		return core.RequestDB{}, err
	}
	return rdb, nil
}
