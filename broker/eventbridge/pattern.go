//
// Copyright (C) 2021 - 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package eventbridge

import (
	"encoding/json"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/swarm"
)

// Event Pattern expression
type Expr func(*awsevents.EventPattern)

// Like combinator defines DSL to configure awsevents.EventPattern
// https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-create-pattern.html
func Like(seq ...Expr) *awsevents.EventPattern {
	p := &awsevents.EventPattern{}
	for _, f := range seq {
		f(p)
	}
	return p
}

// Configure event pattern to match a type
func Type[T any]() Expr {
	return func(p *awsevents.EventPattern) {
		cat := swarm.TypeOf[T]()
		if p.DetailType == nil {
			p.DetailType = jsii.Strings()
		}
		*p.DetailType = append(*p.DetailType, jsii.String(cat))
	}
}

// Configure event pattern to match a category
func Category(cats ...string) Expr {
	return func(p *awsevents.EventPattern) {
		if p.DetailType == nil {
			p.DetailType = jsii.Strings()
		}
		for _, cat := range cats {
			*p.DetailType = append(*p.DetailType, jsii.String(cat))
		}
	}
}

// Configure event pattern to match source(s)
func Source(sources ...string) Expr {
	return func(p *awsevents.EventPattern) {
		if p.Source == nil {
			p.Source = jsii.Strings()
		}
		for _, src := range sources {
			*p.Source = append(*p.Source, jsii.String(src))
		}
	}
}

// Configure event pattern to match object's key
func Has(key string, values ...any) Expr {
	return func(p *awsevents.EventPattern) {
		if p.Detail == nil {
			p.Detail = &map[string]any{}
		}

		var seq []string
		val, has := (*p.Detail)[key]
		if !has {
			seq = []string{}
		} else {
			seq = val.([]string)
		}

		for _, value := range values {
			b, err := json.Marshal(value)
			if err != nil {
				continue
			}
			if b[0] == '"' {
				seq = append(seq, string(b[1:len(b)-1]))
			} else {
				seq = append(seq, string(b))
			}
		}

		(*p.Detail)[key] = seq
	}
}

// Configure event pattern to match Event's Meta key
func EventMeta(key string, values ...any) Expr {
	return Has("meta."+key, values...)
}

// Configure event pattern to match Event's Data key
func EventData(key string, values ...any) Expr {
	return Has("data."+key, values...)
}
