package internal

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type VerificationStrategy interface {
	Verify(string) error
}

type Handler interface {
	Start(chan string) error
	Shutdown()
	Name() string
}

type CommandCenter struct {
	verifier VerificationStrategy
	handlers []Handler
	cmd      []string
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
			err := c.verifier.Verify(read)
			if err == nil {
				log.Info().Msgf("Received and successfully verified message, running cmd %s with args %v", c.cmd[0], c.cmd[1:])
				err := runHook(c.cmd)
				if err != nil {
					log.Error().Err(err)
				}
			} else {
				log.Warn().Msgf("Received message but could not verify it: %v", err)
			}
		}
	}()

	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGINT, syscall.SIGHUP)

	<-done
	close(msgQueue)
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
			log.Error().Msgf("Could not start handler %s: %v", handler.Name(), err)
		}
	}
	return nil
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
