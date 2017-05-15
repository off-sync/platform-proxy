package infraaws

import (
	"context"
	"errors"

	"encoding/json"

	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/off-sync/platform-proxy/app/config/cmd/startwatcher"
)

var (
	ErrQueueNotFound       error = errors.New("Queue not found")
	ErrMultipleQueuesFound error = errors.New("Multiple queues found, expected 1")
)

type SqsStartWatcherCommand struct {
	sqsSvc          *sqs.SQS
	queueUrl        *string
	pollingInterval int
}

func NewSqsStartWatcherCommand(p client.ConfigProvider, queueName string, pollingInterval int) (*SqsStartWatcherCommand, error) {
	sqsSvc := sqs.New(p)

	lqo, err := sqsSvc.ListQueues(&sqs.ListQueuesInput{
		QueueNamePrefix: aws.String(queueName),
	})
	if err != nil {
		return nil, err
	}

	if len(lqo.QueueUrls) < 1 {
		return nil, ErrQueueNotFound
	} else if len(lqo.QueueUrls) > 1 {
		return nil, ErrMultipleQueuesFound
	}

	if pollingInterval < 1 {
		pollingInterval = 1
	}

	return &SqsStartWatcherCommand{
		sqsSvc:          sqsSvc,
		queueUrl:        lqo.QueueUrls[0],
		pollingInterval: pollingInterval,
	}, nil
}

func (c *SqsStartWatcherCommand) Execute(model *startwatcher.CommandModel) error {
	go c.runWatcher(model.WaitGroup, model.Ctx, model.Callback)

	return nil
}

func (c *SqsStartWatcherCommand) runWatcher(wg *sync.WaitGroup, ctx context.Context, cb startwatcher.ChangesCallback) {
	wg.Add(1)

	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			c.pollQueue(ctx, cb)
		}
	}
}

func (c *SqsStartWatcherCommand) pollQueue(ctx aws.Context, cb startwatcher.ChangesCallback) error {
	rmo, err := c.sqsSvc.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:          c.queueUrl,
		VisibilityTimeout: aws.Int64(3),
		WaitTimeSeconds:   aws.Int64(int64(c.pollingInterval)),
	})
	if err != nil {
		return err
	}

	for _, msg := range rmo.Messages {
		// always delete received messages
		c.sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      c.queueUrl,
			ReceiptHandle: msg.ReceiptHandle,
		})

		changes := &startwatcher.Changes{}

		err = json.Unmarshal([]byte(*msg.Body), changes)
		if err == nil {
			cb(changes)
		}
	}

	return nil
}
