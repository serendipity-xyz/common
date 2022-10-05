package sqs_test

import (
	"testing"

	"github.com/serendipity-xyz/common/sqs"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	_, err := sqs.NewClient(&sqs.ClientParams{
		Region:       "us-east-1",
		AccessKey:    "",
		AccessSecret: "aaa",
		QueueName:    "test",
	})
	require.NotNil(t, err, "there should be an err if no access key")
}
