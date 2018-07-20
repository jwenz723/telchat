package tcp

import (
	"github.com/sirupsen/logrus"
	"fmt"
	"net"
)

// Handler contains options for a net.Listener as well as a way to handle all new connections that are accepted
type Handler struct {
	address 			string
	done				chan struct{}
	logger 				*logrus.Logger
	newConnections 		chan net.Conn
	port 				int
	Ready				bool // Indicates that the http listener is ready to accept connections
}

// New will create a new Handler for starting a new TCP listener
func New(address string, port int, newConnections chan net.Conn, logger *logrus.Logger) *Handler {
	return &Handler{
		address:		address,
		done:			make(chan struct{}),
		logger:      	logger,
		newConnections: newConnections,
		port:			port,
	}
}

// Start starts the TCP listener and accepts incoming connections indefinitely until Stop() is called
func (h *Handler) Start() error {
	defer func() {
		h.Ready = false
		close(h.done)
		//h.done = nil
	}()
	//return fmt.Errorf("test")
	// Start the TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", h.address, h.port))
	if err != nil {
		return err
	}
	h.Ready = true
	h.logger.WithFields(logrus.Fields{
		"address": listener.Addr(),
	}).Info("TCP listener accepting connections")

	// pulled this code from the example at: https://stackoverflow.com/a/18969608/3703667
	for {
		type accepted struct {
			conn net.Conn
			err  error
		}
		c := make(chan accepted, 1)
		go func() {
			conn, err := listener.Accept()
			c <- accepted{conn, err}
		}()

		select {
		case a := <-c:
			if a.err != nil {
				h.logger.WithField("error", a.err).Fatal("error accepting connection")
				continue
			}
			h.newConnections <- a.conn
		case <-h.done:
			h.logger.Info("stopping TCP listener...")
			listener.Close()
			return nil
		}
	}
}

// Stop will shutdown the TCP listener
func (h *Handler) Stop() {
	if h.Ready && h.done != nil {
		h.done <- struct {}{}

		// wait for the done channel to be closed (meaning the Start() func has actually stopped running)
		<-h.done
		h.done = nil
	}
}