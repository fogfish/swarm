//
// Copyright (C) 2021 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/broker/sqs"
	"github.com/fogfish/swarm/queue/events"
)

//
// Date type (object) affected by events
type User struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

//
// Events
type EventCreateUser swarm.Event[*User]

func (EventCreateUser) HKT1(swarm.EventType) {}
func (EventCreateUser) HKT2(*User)           {}

type EventUpdateUser swarm.Event[*User]

func (EventUpdateUser) HKT1(swarm.EventType) {}
func (EventUpdateUser) HKT2(*User)           {}

type EventRemoveUser swarm.Event[*User]

func (EventRemoveUser) HKT1(swarm.EventType) {}
func (EventRemoveUser) HKT2(*User)           {}

func main() {
	q, err := sqs.New("swarm-test")
	if err != nil {
		panic(err)
	}

	fmt.Println(xxx[*User, EventCreateUser]())
	fmt.Println(xxx[*Note, swarm.Event[*Note]]())

	userCreated, _ := events.Enqueue[*User, EventCreateUser](q)
	userUpdated, _ := events.Enqueue[*User, EventUpdateUser](q)
	userRemoved, _ := events.Enqueue[*User, EventRemoveUser](q)
	note, _ := events.Enqueue[*Note, swarm.Event[*Note]](q)

	//
	// Multiple channels emits events
	userCreated <- &EventCreateUser{
		Agent:       "example",
		Participant: "user",
		Object:      &User{ID: "user", Text: "some text"},
	}

	userUpdated <- &EventUpdateUser{
		Agent:       "example",
		Participant: "user",
		Object:      &User{ID: "user", Text: "some text with changes"},
	}

	userRemoved <- &EventRemoveUser{
		Agent:       "example",
		Participant: "user",
		Object:      &User{ID: "note"},
	}

	//
	// Single channel emits event
	note <- &swarm.Event[*Note]{
		Type:        "note:EventCreateNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text"},
	}

	note <- &swarm.Event[*Note]{
		Type:        "note:EventUpdateNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note", Text: "some text with changes"},
	}

	note <- &swarm.Event[*Note]{
		Type:        "note:EventRemoveNote",
		Agent:       "example",
		Participant: "user",
		Object:      &Note{ID: "note"},
	}

	q.Close()
}

func xxx[T any, E swarm.EventKind[T]]() string {
	return strings.ToLower(typeOf[T]()) + ":" + typeOf[E]()
}

func typeOf[T any]() string {
	//
	// TODO: fix
	//   Action[*swarm.User] if container type is used
	//

	typ := reflect.TypeOf(*new(T))
	cat := typ.Name()
	if typ.Kind() == reflect.Ptr {
		cat = typ.Elem().Name()
	}

	return cat
}
