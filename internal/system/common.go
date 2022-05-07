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
