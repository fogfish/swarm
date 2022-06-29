//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

//
// The file declares common types used by EventBridge adapter
//

package eventbridge

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

type PutEvents interface {
	PutEvents(context.Context, *eventbridge.PutEventsInput, ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type EnqueueService interface {
	PutEvents
}
