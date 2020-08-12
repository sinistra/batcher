package main

import (
	"log"
	"strconv"
	"time"

	"github.com/sinistra/batcher"
)

// DoSomething is a func that conforms to the type defined in batcher.
// This func is passed into the batcher to do the processing.
func DoSomething(workerID int, records []batcher.Job) []batcher.JobResult {
	var results []batcher.JobResult
	for i, job := range records {
		log.Printf("Processing %d - %s on worker %d\n", i, job.Id, workerID)
		result := batcher.JobResult{
			Id:     job.Id,
			Body:   job.Body,
			Status: "ok",
		}
		results = append(results, result)
	}
	return results
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Go batching")

	batchSize := 10
	batchWait := time.Millisecond * 3000

	// create the batch
	b := batcher.NewBatch(batchSize, batchWait, DoSomething, 1)
	log.Println("batch created")

	// start the batching service
	go b.Worker()
	log.Println("worker created")

	// fill the batcher with data
	for i := 0; i < 100; i++ {
		b.Wg.Add(1)
		go func(j int) {
			result := b.Insert(batcher.Job{
				Id:   strconv.Itoa(j),
				Body: "some text, probably json",
			})
			log.Printf("Inserted record %d with result of %v\f", j, result)
			b.Wg.Done()
		}(i)
	}

	log.Println("all the inserting is done ... waiting for goroutines to finish")
	// wait for everyone to finish
	b.Wg.Wait()
	log.Println("waited until all the inserting is done")

}
