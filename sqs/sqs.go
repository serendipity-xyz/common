package sqs

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/serendipity-xyz/common/log"
)

type Client struct {
	region    string
	sqsClient *sqs.SQS
	queueURL  *string
	queueName string
}

type ClientParams struct {
	Region       string
	AccessKey    string
	AccessSecret string
	QueueName    string
}

func NewClient(params *ClientParams) (*Client, error) {
	if e := os.Getenv("AWS_ACCESS_KEY_ID"); e == "" && params.AccessKey == "" {
		return nil, errors.New("no access key.. set AWS_ACCESS_KEY_ID environment variable")
	}
	if e := os.Getenv("AWS_SECRET_ACCESS_KEY"); e == "" && params.AccessSecret == "" {
		return nil, errors.New("no access key.. set AWS_SECRET_ACCESS_KEY environment variable")
	}
	creds := credentials.NewStaticCredentials(params.AccessKey, params.AccessSecret, "")
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(params.Region),
	}))
	sqsClient := sqs.New(sess)
	result, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &params.QueueName,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		region:    params.Region,
		sqsClient: sqsClient,
		queueURL:  result.QueueUrl,
		queueName: params.QueueName,
	}, nil
}

func (c *Client) Producer() *Producer {
	return &Producer{
		client: c,
	}
}

func (c *Client) Consumer() *Consumer {
	return &Consumer{
		client:  c,
		msgChan: make(chan *Msg, 1),
	}
}

type Producer struct {
	client *Client
}

type Msg map[string]interface{}

func (m Msg) I(key string) int {
	wrapper, ok := m[key]
	if !ok {
		return -1
	}
	i, ok := wrapper.(int)
	if !ok {
		return -1
	}
	return i
}

func (m Msg) I64(key string) int64 {
	wrapper, ok := m[key]
	if !ok {
		return -1
	}
	i, ok := wrapper.(int64)
	if !ok {
		return -1
	}
	return i
}

func (m Msg) F64(key string) float64 {
	wrapper, ok := m[key]
	if !ok {
		return -1
	}
	f, ok := wrapper.(float64)
	if !ok {
		return -1
	}
	return f
}

func (m Msg) S(key string) string {
	wrapper, ok := m[key]
	if !ok {
		return key
	}
	s, ok := wrapper.(string)
	if !ok {
		return key
	}
	return s
}

func (m Msg) String() (string, error) {
	attempt := m.I("attempt")
	if attempt < 0 {
		m["attempt"] = 0
	}
	res, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (p *Producer) ProduceMsg(msg Msg) (string, error) {
	s, err := msg.String()
	if err != nil {
		return "", err
	}
	res, err := p.client.sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    p.client.queueURL,
		MessageBody: aws.String(s),
	})
	if err != nil {
		return "", err
	}
	return *res.MessageId, nil
}

type Consumer struct {
	client  *Client
	msgChan chan *Msg
}

func (c *Consumer) MsgChan() chan *Msg { return c.msgChan }

func (c *Consumer) Poll(l log.Logger) {
	for {
		l.Info("polling message queue [%v]....", c.client.queueName)
		output, err := c.client.sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            c.client.queueURL,
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(15),
		})
		if err != nil {
			l.Error("failed to fetch sqs message %v", err)
			continue
		}
		for _, sqsMsg := range output.Messages {
			msg := &Msg{
				"messageId":     *sqsMsg.MessageId,
				"receiptHandle": *sqsMsg.ReceiptHandle,
			}
			if err := json.Unmarshal([]byte(*sqsMsg.Body), &msg); err != nil {
				l.Error("failed to fetch sqs message %v", err)
				continue
			}
			c.msgChan <- msg
		}

	}

}

func (c *Consumer) MarkProcessed(l log.Logger, msg *Msg) {
	rh := msg.S("receiptHandle")
	_, err := c.client.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      c.client.queueURL,
		ReceiptHandle: aws.String(rh),
	})
	if err != nil {
		l.Error("failed to delete message [%v]: %v", rh, err)
	}
}
