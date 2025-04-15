package tests

import (
	"os"
	"testing"
	"time"

	"github.com/wisp167/pvz/internal/server"
)

var app *server.Application

func setup() (*server.Application, error) {
	app, err := server.SetupApplication()
	if err != nil {
		return nil, err
	}

	go func() {
		if err := app.Start(); err != nil {
			panic(err)
		}
	}()

	time.Sleep(1 * time.Second)
	return app, nil
}

func teardown(app *server.Application) {
	if err := app.Stop(); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	var err error
	app, err = setup()
	if err != nil {
		panic(err)
	}

	code := m.Run()
	teardown(app)
	time.Sleep(1 * time.Second)
	os.Exit(code)
}
