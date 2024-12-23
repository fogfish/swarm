//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge_test

import (
	"testing"

	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/curie"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm/broker/eventbridge"
)

func TestLike(t *testing.T) {
	t.Run("Type", func(t *testing.T) {
		type Event string

		p := eventbridge.Like(
			eventbridge.Type[Event](),
		)

		it.Then(t).Should(
			it.Seq(*p.DetailType).Equal(jsii.String("Event")),
		)
	})

	t.Run("Category", func(t *testing.T) {
		p := eventbridge.Like(
			eventbridge.Category("Event"),
		)

		it.Then(t).Should(
			it.Seq(*p.DetailType).Equal(jsii.String("Event")),
		)
	})

	t.Run("Source", func(t *testing.T) {
		p := eventbridge.Like(
			eventbridge.Source("source"),
		)

		it.Then(t).Should(
			it.Seq(*p.Source).Equal(jsii.String("source")),
		)
	})

	t.Run("Has", func(t *testing.T) {
		p := eventbridge.Like(
			eventbridge.Has("a.b", curie.IRI("abc"), curie.IRI("def")),
			eventbridge.Has("a.c", "abc", "def"),
			eventbridge.Has("a.d", 10, 11),
			eventbridge.Has("a.e", 10.1, 11.1),
		)

		it.Then(t).Should(
			it.Seq((*p.Detail)["a.b"].([]string)).Equal("[abc]", "[def]"),
			it.Seq((*p.Detail)["a.c"].([]string)).Equal("abc", "def"),
			it.Seq((*p.Detail)["a.d"].([]string)).Equal("10", "11"),
			it.Seq((*p.Detail)["a.e"].([]string)).Equal("10.1", "11.1"),
		)
	})

	t.Run("EventMeta", func(t *testing.T) {
		p := eventbridge.Like(
			eventbridge.EventMeta("agent", curie.IRI("abc")),
		)

		it.Then(t).Should(
			it.Seq((*p.Detail)["meta.agent"].([]string)).Equal("[abc]"),
		)
	})

	t.Run("EventData", func(t *testing.T) {
		p := eventbridge.Like(
			eventbridge.EventData("agent", curie.IRI("abc")),
		)

		it.Then(t).Should(
			it.Seq((*p.Detail)["data.agent"].([]string)).Equal("[abc]"),
		)
	})

}
