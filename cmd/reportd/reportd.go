package main

import (
	"os"
	"time"

	zmq "github.com/pebbe/zmq4"
	"github.com/untangle/packetd/services/logger"
)

var routineWatcher = make(chan int)

// main is the entrypoint of reportd
func main() {

	logger.Startup()

	logger.Info("Starting up reportd...\n")

	logger.Info("Setting up zmq listening socket...\n")
	socket, err := setupZmqSocket()

	if err != nil {
		logger.Warn("Unable to setup ZMQ sockets.")
	}
	defer socket.Close()

	logger.Info("Setting up event listener on zmq socket...\n")
	go eventListener(socket)

	go eventLogger()

	logger.Info("Waiting for all routines to finish...\n")

	numGoroutines := 0
	for diff := range routineWatcher {
		numGoroutines += diff
		logger.Info("Running routines: %v\n", numGoroutines)
		if numGoroutines == 0 {
			logger.Info("Shutting down reportd...\n")
			os.Exit(0)
		}
	}
}

// eventListener is used to listen for ZMQ events being published
func eventListener(soc *zmq.Socket) {
	routineWatcher <- +1
	logger.Info("Starting up a theoretical goroutine for event listening...\n")
	defer startFinishRoutineThread()

	for start := time.Now(); time.Since(start) > 30*time.Second; {
		msg, err := soc.RecvMessage(0)
		if err != nil {
			logger.Warn("Unable to receive messages: %s...\n", err)
			break
		}

		logger.Info("Got some data here, topic: %s, message: %s\n", msg[0], msg[1])
	}

	logger.Info("Shutting down eventListener goroutine...\n")
}

// eventLogger is used to log events to the database
func eventLogger() {
	routineWatcher <- +1
	defer startFinishRoutineThread()

	logger.Info("Starting up a theoretical goroutine for event logging...\n")
	time.Sleep(30 * time.Second)
}

// startFinishRoutineThread is a function to simplify how we can defer calling finishRoutine() at the top of a function,
// instead of having to always call it at the end of a routine
func startFinishRoutineThread() {
	go finishRoutine()
}

// finishRoutine is called at the end of a running go routine to empty the channel watcher
func finishRoutine() {
	routineWatcher <- -1
}

// setupZmqSocket prepares a zmq socket for listening to events
func setupZmqSocket() (soc *zmq.Socket, err error) {
	subscriber, err := zmq.NewSocket(zmq.SUB)

	if err != nil {
		logger.Err("Unable to open ZMQ socket... %s\n", err)
		return nil, err
	}

	subscriber.Connect("tcp://localhost:5561")
	subscriber.SetSubscribe("packetd-events")

	return subscriber, nil
}
