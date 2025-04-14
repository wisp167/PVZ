package server

import (
	"runtime"
	"strings"
)

func (app *Application) logError(err error) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	_, file1, line1, ok1 := runtime.Caller(3)
	if !ok1 {
		file1 = "unknown"
		line = 0
	}

	fileParts := strings.Split(file, "/")
	filename := fileParts[len(fileParts)-1]
	fileParts = strings.Split(file1, "/")
	filename1 := fileParts[len(fileParts)-1]

	app.logger.Printf("[%s:%d]->[%s:%d] %v", filename, line, filename1, line1, err)
}
