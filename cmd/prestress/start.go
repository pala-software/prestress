package main

func start() (err error) {
	c, err := container()
	if err != nil {
		return
	}

	err = c.Invoke(startHttpServer)
	if err != nil {
		return
	}

	return
}
