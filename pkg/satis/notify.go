package satis

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func Notify(topicARN, message string) error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	input := sns.PublishInput{
		Message:  &message,
		TopicArn: &topicARN,
	}

	svc := sns.New(sess)
	_, err := svc.Publish(&input)
	if err != nil {
		return err
	}

	return nil
}
