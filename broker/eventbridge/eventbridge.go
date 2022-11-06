package eventbridge

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/fogfish/swarm"
)

type client struct {
	bus     string
	service EventBridge
}

func newClient(service EventBridge, bus string) (*client, error) {
	api, err := newService(service)
	if err != nil {
		return nil, err
	}

	return &client{
		bus:     bus,
		service: api,
	}, nil
}

func newService(service EventBridge) (EventBridge, error) {
	if service != nil {
		return service, nil
	}

	aws, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return eventbridge.NewFromConfig(aws), nil
}

// Enq enqueues message to broker
func (cli *client) Enq(bag swarm.Bag) error {
	ret, err := cli.service.PutEvents(
		context.TODO(),
		&eventbridge.PutEventsInput{
			Entries: []types.PutEventsRequestEntry{
				{
					EventBusName: aws.String(cli.bus),
					Source:       aws.String(bag.Queue),
					DetailType:   aws.String(bag.Category),
					Detail:       aws.String(string(bag.Object)),
				},
			},
		},
	)
	if err != nil {
		return err
	}

	if ret.FailedEntryCount > 0 {
		return fmt.Errorf("%v: %v", ret.Entries[0].ErrorCode, ret.Entries[0].ErrorMessage)
	}

	return nil
}
