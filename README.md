# batcher

Simple batch library for Golang.
Batcher will collect 'jobs' until batchSize jobs are received and then execute them all in a batch.
A timeout is passed to the batch creator. If no 'jobs' are received within the timeout, the batcher will execute a batch, if there are any jobs to execute.
Batcher will echo 'timout' to the console if there are no jobs to execute.

# motivation & learnings
This library uses go-routines & concurrency extensively.
Part of the motivation for this is to explore the challenges presented by concurrency.
Conceptually, it is a bit more challenging, particularly keeping the queuing of channels well defined.

Still to do - testing - much more testing.

Conclusions:
- Concurrency makes testing considerably more difficult.
- Use concurrency sparingly.

## How to Use

Install:

```
go get github.com/sinistra/batcher
```

Example:

There is an example of how to use this library in the example folder.


step #1:
create a new batch
```
b := batcher.NewBatch(batchSize, batchWait, DoSomething, 1)
```
batchSize is an integer of the max size of each batch.  
batchWait is the max time duration between batches.  
DoSomething is a function that is provided by the caller. This function will be executed by the batch process against each job. This function must conform with the type defined in batcher. An example function is provided in the example program. It must return a slice of batcher.JobResult  

```
func DoSomething(workerID int, records []batcher.Job) []batcher.JobResult {
    //do something
    return []batcher.JobResult
}
```

step #2

call the batch Worker
```
go b.Worker()
```
This should be called as a goroutine so that it runs concurrently.

step #3

submit records to the batcher by calling b.Insert().  
The only parameter passed to insert is a struct of type Job.  
The call to Insert should wrapped inside a goroutine. This is because the Insert function is looking for a response from the batch worker. We need the worker to be peppered with a number of concurrent Inserts so that it will trigger a batch execution.  

IF the Insert is not wrapped in a goroutine, this has the effect of sequentially inserting 1 record at a time and waiting for the timeout to trigger a batch of 1. Processing performance of this scenario will be worse than sequentially processing the jojbs without the assistance of the batcher.


