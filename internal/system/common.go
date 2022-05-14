//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package system

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

//
//
func NewSession(defSession ...*session.Session) (*session.Session, error) {
	if len(defSession) != 0 {
		return defSession[0], nil
	}

	awscli, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	return awscli, nil
}
