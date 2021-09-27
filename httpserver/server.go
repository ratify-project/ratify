package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"

	ef "github.com/deislabs/hora/pkg/executor/core"
	"github.com/gorilla/mux"
)

const (
	ServerRootURL = "/hora/v1"
)

type (
	Server struct {
		Address  string
		Router   *mux.Router
		Executor *ef.Executor
		Context  context.Context
	}
)

func NewServer(context context.Context, address string, executor *ef.Executor) (*Server, error) {
	if address == "" {
		return nil, ServerAddrNotFoundError{}
	}

	server := &Server{
		Address:  address,
		Executor: executor,
		Router:   mux.NewRouter(),
		Context:  context,
	}
	server.registerHandlers()

	return server, nil
}

func (server *Server) Run() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", server.Address)
	if err != nil {
		return err
	}
	lsnr, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	svr := http.Server{
		Addr:    server.Address,
		Handler: server.Router,
	}
	return svr.Serve(lsnr)
}

func (server *Server) register(method, path string, handler ContextHandler) {
	server.Router.Methods(method).Path(path).Handler(&contextHandler{
		context: server.Context,
		handler: handler,
	})
}

func (server *Server) registerHandlers() {
	server.register("GET", ServerRootURL+"/verify", server.verify)
}

type ServerAddrNotFoundError struct{}

func (err ServerAddrNotFoundError) Error() string {
	return fmt.Sprint("The http server address configuration is not set. Skipping server creation")
}
