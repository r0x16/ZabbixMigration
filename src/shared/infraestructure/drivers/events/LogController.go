package events

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
)

type LogController struct {
	path string
}

var eventLogMutex sync.Mutex

func NewLogController(path string) (*LogController, *model.Error) {
	file, openError := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if openError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: openError.Error(),
		}
	}
	defer file.Close()
	return &LogController{
		path: path,
	}, nil
}

func (controller *LogController) GetCurrentLog() ([]string, *model.Error) {
	eventLogMutex.Lock()
	defer eventLogMutex.Unlock()

	file, openError := os.OpenFile(controller.path, os.O_RDONLY|os.O_CREATE, 0644)
	if openError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: openError.Error(),
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var text []string

	for scanner.Scan() {
		text = append(text, scanner.Text()+"\n")
	}

	return text, nil
}

func (controller *LogController) GetLogFromLine(from int) ([]string, *model.Error) {
	eventLogMutex.Lock()
	defer eventLogMutex.Unlock()

	file, openError := os.OpenFile(controller.path, os.O_RDONLY|os.O_CREATE, 0644)
	if openError != nil {
		return nil, &model.Error{
			Code:    http.StatusInternalServerError,
			Message: openError.Error(),
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var text []string
	count := 0

	for scanner.Scan() {
		if count >= from {
			text = append(text, scanner.Text()+"\n")
		}
		count++
	}

	return text, nil
}

func (controller *LogController) WriteLog(text string) (string, *model.Error) {
	eventLogMutex.Lock()
	defer eventLogMutex.Unlock()

	file, openError := os.OpenFile(controller.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if openError != nil {
		return "", &model.Error{
			Code:    http.StatusInternalServerError,
			Message: openError.Error(),
		}
	}
	defer file.Close()

	strLog := &strings.Builder{}
	multiLog := io.MultiWriter(file, strLog)

	log.SetOutput(multiLog)
	log.Println(text)

	return strLog.String(), nil
}

func (controller *LogController) Path() string {
	return controller.path
}
