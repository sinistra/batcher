package batcher

import (
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"testing"
	"time"
)

func DoSomething(workerID int, records []Job) []JobResult {
	var results []JobResult
	for i, job := range records {
		log.Printf("Processing %d - %s on worker %d\n", i, job.Id, workerID)
		result := JobResult{
			Id:     job.Id,
			Body:   job.Body,
			Status: "ok",
		}
		results = append(results, result)
	}
	return results
}

func TestBatch_Insert(t *testing.T) {
	type args struct {
		job Job
	}
	tests := []struct {
		name string
		args args
		want JobResult
	}{
		{
			name: "insert test",
			args: args{job: Job{
				Id:   "1",
				Body: "some text",
			}},
			want: JobResult{
				Id:     "1",
				Body:   "some text",
				Status: "ok",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchWait := time.Millisecond * 3000
			b := NewBatch(10, batchWait, DoSomething, 1)
			go b.Worker()
			if got := b.Insert(tt.args.job); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Insert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatch_Quit(t *testing.T) {
	batchWait := time.Millisecond * 3000
	b := NewBatch(10, batchWait, DoSomething, 1)
	go b.Worker()
	assert.True(t, b.Quit())
}

func TestNewBatch(t *testing.T) {
	batchWait := time.Millisecond * 3000
	b := NewBatch(10, batchWait, DoSomething, 1)
	assert.IsType(t, &Batch{}, b)
}
