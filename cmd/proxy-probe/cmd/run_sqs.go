package cmd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spf13/viper"
)

var (
	sqsQueueName string
)

func init() {
	runCmd.PersistentFlags().StringVarP(&ecsClusterName, "sqs-queue-name", "q", "",
		"SQS queue on which frontend & backend updates are posted")
	viper.BindPFlag("sqsQueueName", runCmd.PersistentFlags().Lookup("sqs-queue-name"))
}

func checkSqs(p client.ConfigProvider) {
	sqsSvc := sqs.New(p)

	queueName := viper.GetString("sqsQueueName")

	// SQS::ListQueues
	le := log.WithField("queue_name", queueName)

	lqo, err := sqsSvc.ListQueues(&sqs.ListQueuesInput{
		QueueNamePrefix: &queueName,
	})
	if err != nil {
		le.WithError(err).Fatal("SQS::ListQueues failed")
	}

	if len(lqo.QueueUrls) != 1 {
		le.WithError(err).Fatal("SQS::ListQueues failed: queue not found")
	}

	queueURL := *lqo.QueueUrls[0]

	le.WithField("queue_url", queueURL).Info("SQS::ListQueues successful")

	// SQS::ReceiveMessage
	le = log.WithField("queue_url", queueURL)

	rmo, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		VisibilityTimeout:   aws.Int64(3),
		WaitTimeSeconds:     aws.Int64(10),
		MaxNumberOfMessages: aws.Int64(1),
	})
	if err != nil {
		le.WithError(err).Fatal("SQS::ReceiveMessage failed")
	}

	if len(rmo.Messages) != 1 {
		le.WithError(err).Fatal("SQS::ReceiveMessage failed: no messages available")
	}

	receiptHandle := *rmo.Messages[0].ReceiptHandle

	le.WithField("receipt_handle", receiptHandle).Info("SQS::ReceiveMessage successful")

	// SQS::DeleteMessage
	le = log.WithField("receipt_handle", receiptHandle)

	_, err = sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: &receiptHandle,
	})
	if err != nil {
		le.WithError(err).Fatal("SQS::DeleteMessage failed")
	}

	le.Info("SQS::DeleteMessage successful")
}
