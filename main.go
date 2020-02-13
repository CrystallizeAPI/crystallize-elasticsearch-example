package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/crystallizeapi/crystallize-elasticsearch-example/tasks"
)

func runTask(taskName *string, tenant *string) error {
	ctx := context.Background()

	// Create the task based on the name
	task, err := tasks.NewTask(*taskName, *tenant)
	if err != nil {
		return err
	}

	// Setup and execute the task
	fmt.Printf("Running task %s\n", *taskName)
	begin := time.Now()
	if err := task.Setup(ctx); err != nil {
		return err
	}
	if err := task.Execute(ctx); err != nil {
		return err
	}
	fmt.Printf("Task completed in %f seconds\n", time.Since(begin).Seconds())

	return nil
}

func runServer() {
	// http.HandleFunc("/search", server.ServeHTTP)

	// http.ListenAndServe(":8090", nil)
}

func main() {
	var (
		mode     = flag.String("mode", "server", "mode for the application (task|server)")
		taskName = flag.String("task", "", "task to run")
		tenant   = flag.String("tenant", "", "tenant identifier")
	)
	flag.Parse()

	if *mode == "task" {
		err := runTask(taskName, tenant)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		runServer()
	}
}
