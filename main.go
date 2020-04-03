package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CrystallizeAPI/crystallize-elasticsearch-example/server"
	"github.com/CrystallizeAPI/crystallize-elasticsearch-example/tasks"
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
	http.HandleFunc("/api/index", server.HandleIndex)
	http.HandleFunc("/api/search", server.HandleSearch)

	fmt.Printf("Listening on http://localhost:8090\n")
	http.ListenAndServe(":8090", nil)
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
