package internal

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

// VerificationStrategy allows defining multiple strategies to either drop or accept an incoming message.
type VerificationStrategy interface {
	// Verify accepts a received message and determines whether the message is ought to be accepted or not. Returns an
	// error if the verification failed and the message should be dropped.
	Verify(string) error
}

// Handler receives shutdown messages from different sources and dispatches them to the command center.
type Handler interface {
	// Start initializes the handler in order to start receiving messages. It accepts the channel that it should
	// write the received messages to.
	Start(chan string) error

	// Shutdown shuts the handler down safely.
	Shutdown()

	// Name returns the distinctive name of this handler.
	Name() string
}

type CommandCenter struct {
	// selected strategy to verify the received messages
	verifier VerificationStrategy

	// slice of handlers that we can receive messages from
	handlers []Handler

	// the cmd to execute once a received message has been accepted
	cmd []string
}

func NewCommandCenter(verification VerificationStrategy, handlers []Handler, cmd []string) (*CommandCenter, error) {
	if nil == verification {
		return nil, errors.New("no verificationstrategy provided")
	}

	if nil == handlers || len(handlers) == 0 {
		return nil, errors.New("no handlers provided")
	}

	if nil == cmd || len(cmd) == 0 {
		return nil, errors.New("no cmd provided")
	}

	return &CommandCenter{verifier: verification, handlers: handlers, cmd: cmd}, nil
}

func (c *CommandCenter) Start() error {
	msgQueue := make(chan string)
	err := c.startHandlers(msgQueue)
	if err != nil {
		return err
	}

	go func() {
		for read := range msgQueue {
			verifyAndRun(c.verifier, read, c.cmd)
		}
	}()

	heartbeat := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			<-heartbeat.C
			MetricHeartbeat.SetToCurrentTime()
		}
	}()

	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGINT, syscall.SIGHUP)

	<-done
	close(msgQueue)
	heartbeat.Stop()
	log.Info().Msgf("Received signal, shutting down...")
	for _, handler := range c.handlers {
		handler.Shutdown()
	}

	return nil
}

func (c *CommandCenter) startHandlers(msgQueue chan string) error {
	for _, handler := range c.handlers {
		log.Info().Msgf("Starting handler %s", handler.Name())
		err := handler.Start(msgQueue)
		if err != nil {
			return fmt.Errorf("could not start handler %s: %v", handler.Name(), err)
		}
	}
	return nil
}

func verifyAndRun(verifier VerificationStrategy, read string, commands []string) {
	err := verifier.Verify(read)
	if err == nil {
		log.Info().Msgf("Received and successfully verified message, running cmd %s with args %v", commands[0], commands[1:])
		err := runHook(commands)
		if err != nil {
			log.Error().Msgf("Could not run command: %v", err)
			MetricCommandExecutionFailures.Inc()
		}
	} else {
		MetricMessageVerifyErrors.Inc()
		log.Warn().Msgf("Received message but could not verify it: %v", err)
	}
}

func runHook(command []string) error {
	if nil == command || len(command) == 0 {
		return errors.New("empty command given")
	}

	cmd := exec.Command(command[0], command[1:]...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not run command: %v", err)
	}
	return nil
}
