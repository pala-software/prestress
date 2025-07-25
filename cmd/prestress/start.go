package main

import (
	"fmt"
)

func start() (err error) {
	c, err := container()
	if err != nil {
		return
	}

	err = c.Invoke(startHttpServer)
	if err != nil {
		return
	}

	fmt.Println("Server started!")
	return
}
