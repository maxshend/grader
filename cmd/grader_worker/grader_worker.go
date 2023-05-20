package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const workersCount = 5

type SubmissionTask struct {
	GraderURL    string            `json:"grader_url"`
	WebhookURL   string            `json:"webhook_url"`
	Container    string            `json:"container"`
	PartID       string            `json:"part_id"`
	Files        []*SubmissionFile `json:"files"`
	SubmissionID int64             `json:"submission_id"`
}

type SubmissionFile struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if len(rabbitURL) == 0 {
		log.Fatal("RABBITMQ_URL should be set")
	}
	rabbitQueueName := os.Getenv("RABBITMQ_QUEUE")
	if len(rabbitQueueName) == 0 {
		log.Fatal("RABBITMQ_QUEUE should be set")
	}
	rabbitConn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()
	rabbitCh, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitCh.Close()

	_, err = rabbitCh.QueueDeclare(
		rabbitQueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	err = rabbitCh.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatal(err)
	}

	tasks, err := rabbitCh.Consume(
		rabbitQueueName, // queue
		"",              // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(workersCount)

	for i := 0; i <= workersCount; i++ {
		go submitWorker(wg, tasks)
	}

	log.Println("Grader Worker started...")
	wg.Wait()
}

func submitWorker(wg *sync.WaitGroup, tasks <-chan amqp.Delivery) {
	defer wg.Done()
	for taskItem := range tasks {
		log.Printf("Incoming Task: %+v\n", taskItem)

		func(taskItem amqp.Delivery) {
			defer taskItem.Ack(false)

			task := &SubmissionTask{}
			err := json.Unmarshal(taskItem.Body, task)
			if err != nil {
				log.Printf("Can't unpack json: %q\n", err)
				return
			}

			response, err := sendGraderRequest(task.GraderURL, taskItem.Body)
			if err != nil {
				log.Printf(
					"Error while sending submission #%d to the grader (%s): %q\n",
					task.SubmissionID,
					task.GraderURL,
					err,
				)
				return
			}

			responseBody, err := io.ReadAll(response.Body)
			if err != nil {
				log.Fatalln(err)
			}

			log.Printf(
				"Grader response (%s) for submission #%d status %d\n%s\n",
				task.GraderURL,
				task.SubmissionID,
				response.StatusCode,
				responseBody,
			)
		}(taskItem)
	}
}

func sendGraderRequest(posturl string, body []byte) (*http.Response, error) {
	request, err := http.NewRequest("POST", posturl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Minute * 5,
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
