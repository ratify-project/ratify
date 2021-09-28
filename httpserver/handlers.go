package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	e "github.com/deislabs/hora/pkg/executor"
	"github.com/sirupsen/logrus"
)

func (server *Server) verify(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	logrus.Infof("start request %v %v", r.Method, r.URL)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	subject := string(body)

	logrus.Infof("subject for request %v %v is %v", r.Method, r.URL, subject)

	verifyParameters := e.VerifyParameters{
		Subject: subject,
	}

	result, err := server.Executor.VerifySubject(ctx, verifyParameters)

	if err != nil {
		return err
	}

	logrus.Infof("request %v %v completed successfully", r.Method, r.URL)
	res, err := json.MarshalIndent(result, "", "  ")
	if err == nil {
		fmt.Println(string(res))
	}

	if result.IsSuccess {
		return serveJSON(w, "true")
	} else {
		return serveJSON(w, "false")
	}
}

func serveJSON(w http.ResponseWriter, result string) error {
	if err := json.NewEncoder(w).Encode(result); err != nil {
		return err
	}
	return nil
}
