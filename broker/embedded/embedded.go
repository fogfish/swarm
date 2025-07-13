//
// Copyright (C) 2021 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package embedded

import (
	"context"
	"sync"

	"github.com/fogfish/guid/v2"
	"github.com/fogfish/swarm"
)

type Client struct {
	config swarm.Config

	// Control-plane for managing I/O loop
	context context.Context
	cancel  context.CancelFunc
	once    sync.Once

	// Channels + inflight buffer
	bags map[string]*swarm.Bag
	recv <-chan *swarm.Bag
	emit chan<- *swarm.Bag
}

func (cli *Client) Close() error {
	cli.once.Do(cli.cancel)
	return nil
}

func (cli *Client) Enq(ctx context.Context, bag swarm.Bag) error {
	bag.Digest = guid.G(guid.Clock).String()

	select {
	case cli.emit <- &bag:
		return nil
	case <-ctx.Done():
		return swarm.ErrServiceIO
	}
}

func (cli *Client) Ack(ctx context.Context, digest string) error {
	delete(cli.bags, digest)
	return nil
}

func (cli *Client) Err(ctx context.Context, digest string, err error) error {
	if bag, has := cli.bags[digest]; has {
		delete(cli.bags, digest)

		select {
		case cli.emit <- bag:
			return nil
		case <-ctx.Done():
			return swarm.ErrServiceIO
		}

	}
	return nil
}

func (cli *Client) Ask(ctx context.Context) ([]swarm.Bag, error) {
	req, cancel := context.WithTimeout(context.Background(), cli.config.NetworkTimeout*2)
	defer cancel()

	select {
	case bag := <-cli.recv:
		cli.bags[bag.Digest] = bag
		return []swarm.Bag{*bag}, nil
	case <-req.Done():
		return nil, nil
	case <-ctx.Done():
		return nil, swarm.ErrServiceIO
	}
}
