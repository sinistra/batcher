package batcher

import (
	"log"
	"sync"
	"time"
)

// Job is the structure of all messages sent for processing
type Job struct {
	Id   string `json:"id"`
	Body string `json:"body"`
}

// JobResult is the structure of all messages sent for processing
type JobResult struct {
	Id     string `json:"id"`
	Body   string `json:"body"`
	Status string `json:"status"`
}

// BatchFn is a type defn. All processors passed to the batcher must conform to this type
type BatchFn func(workerID int, records []Job) []JobResult

// Batch is the object that controls the batching process
type Batch struct {
	maxSize    int
	maxWait    time.Duration
	incoming   chan Job
	processing chan JobResult
	quit       chan bool
	doneItems  map[string]JobResult
	callback   BatchFn
	Wg         sync.WaitGroup
	items      []interface{}
	mutex      *sync.RWMutex
}

// NewBatch creates a new Batch and returns the object to the caller
func NewBatch(batchSize int, batchWait time.Duration, callback BatchFn, workerCount int) *Batch {
	batch := &Batch{
		maxSize:    batchSize,
		maxWait:    batchWait,
		incoming:   make(chan Job, batchSize),
		processing: make(chan JobResult, batchSize),
		quit:       make(chan bool),
		doneItems:  make(map[string]JobResult, batchSize),
		callback:   callback,
		items:      make([]interface{}, batchSize),
		mutex:      &sync.RWMutex{},
	}
	log.Printf("maxSize is %d\n", batch.maxSize)
	log.Printf("maxWait is %v\n", batch.maxWait)
	return batch
}

// Worker is called from a batch object to control the processing of the batch
func (b *Batch) Worker() {
	b.Wg.Add(1)
	defer b.Wg.Done()

	var jobs []Job
	processed := 0
	timeout := time.After(b.maxWait)
	for {
		select {
		case job, ok := <-b.incoming:
			log.Println("job from incoming")
			if !ok {
				log.Println("Worker shutting down")
				if len(jobs) > 0 {
					log.Printf("Sending final batch of %d jobs\n", len(jobs))
					b.processJobs(jobs)
				}
				return
			}

			jobs = append(jobs, job)
			processed++
			log.Printf("there are %d jobs\n", len(jobs))

			if len(jobs) == b.maxSize {
				log.Println("worker sent a batch")
				b.processJobs(jobs)
				// empty the job slice when the jobs have been processed
				jobs = jobs[:0]
				timeout = time.After(b.maxWait)
			} else {
				log.Println("jobs is not big enough yet")
			}
		case processedJob, ok := <-b.processing:
			if !ok {
				log.Println("Something not working on processed Q")
			}
			b.mutex.Lock()
			log.Printf("Adding job %s to processed list\n", processedJob.Id)
			b.doneItems[processedJob.Id] = processedJob
			b.mutex.Unlock()
		case <-b.quit:
			log.Println("got the quit signal")
			close(b.incoming)
			if len(jobs) > 0 {
				log.Printf("Sending final batch of %d jobs\n", len(jobs))
				b.processJobs(jobs)
			} else {
				log.Printf("We're quitting .. %d jobs\n", len(jobs))
			}
			return
		case <-timeout:
			log.Print("timeout occurred, looking for jobs")
			if len(jobs) > 0 {
				b.processJobs(jobs)
				jobs = jobs[:0]
				log.Println("worker sent a partial batch")
			}
			timeout = time.After(b.maxWait)
		default:
			// log.Println("no activity")
		}
	}
}

// processJobs calls the func passed to the batcher to actually do the work
func (b *Batch) processJobs(jobs []Job) {
	results := b.callback(1, jobs)
	for _, result := range results {
		select {
		case b.processing <- result:
			log.Printf("job %s sent to processing\n", result.Id)
		default:
			log.Println("no job sent")
		}
	}
}

// Insert adds a new job to the batching process
func (b *Batch) Insert(job Job) JobResult {
	log.Println("add job to incoming channel")
	select {
	case b.incoming <- job:
		log.Printf("job %s sent to incoming\f", job.Id)
	default:
		// log.Println("no job sent")
	}
	for {
		b.mutex.Lock()
		result, ok := b.doneItems[job.Id]
		b.mutex.Unlock()
		if ok {
			log.Printf("%s is done.\n", job.Id)
			return result
			break
			// } else {
			//     log.Println("key not found ... waiting")
		}
	}
	// will never get here - just for the compiler
	return JobResult{}
}

// Quit shuts the batcher down
func (b *Batch) Quit() bool {
	b.quit <- true
	log.Println("Sending quit signal")
	return true
}
