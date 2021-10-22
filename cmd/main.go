package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/shutdown-listener/internal"
	"github.com/soerenschneider/shutdown-listener/internal/config"
	mqtt2 "github.com/soerenschneider/shutdown-listener/internal/handlers/mqtt"
	"github.com/soerenschneider/shutdown-listener/internal/verification"
	"os"
	"os/user"
	"path"
	"strings"
)

const (
	cliConfFile = "conf"
	cliVersion  = "version"
)

var envConfFile = fmt.Sprintf("%s_CONFIG", strings.ToUpper(config.AppName))

func main() {
	configFile := ParseCliFlags()
	conf, err := config.ReadJsonConfig(configFile)
	if err != nil {
		log.Fatal().Msgf("could not read config file from %s: %v", err)
	}

	verificationStrategy, err := buildVerification(conf)
	if err != nil {
		log.Fatal().Msgf("could not build verification strategy: %v", err)
	}

	handlers, err := buildHandlers(conf)
	if err != nil {
		log.Fatal().Msgf("could not build handlers: %v", err)
	}

	log.Info().Msg("Building command center..")
	cmd, err := internal.NewCommandCenter(verificationStrategy, handlers, conf.Command)
	if err != nil {
		log.Fatal().Msgf("could not build command center: %v", err)
	}

	cmd.Start()
}

func buildVerification(conf *config.Config) (internal.VerificationStrategy, error) {
	log.Info().Msg("Building message verification implementation...")
	// TODO: implement
	return &verification.NoTrust{}, nil
}

func buildHandlers(conf *config.Config) ([]internal.Handler, error) {
	// TODO: Implement
	log.Info().Msg("Building message handlers...")
	handlers := make([]internal.Handler, 0)
	mqtt, err := mqtt2.NewMqttHandler(&conf.MqttConfig)
	if err != nil {
		return nil, fmt.Errorf("could not build mqtt handler: %v", err)
	}

	handlers = append(handlers, mqtt)

	return handlers, nil
}

func ParseCliFlags() (configFile string) {
	flag.StringVar(&configFile, cliConfFile, os.Getenv(envConfFile), "path to the config file")
	version := flag.Bool(cliVersion, false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("%s (revision %s)", internal.BuildVersion, internal.CommitHash)
		os.Exit(0)
	}
	log.Info().Msgf("This is %s version %s, commit %s", config.AppName, internal.BuildVersion, internal.CommitHash)

	if len(configFile) == 0 {
		log.Fatal().Msgf("No config file specified, use flag '-%s' or env var '%s'", cliConfFile, envConfFile)
	}

	if strings.HasPrefix(configFile, "~/") {
		configFile = path.Join(getUserHomeDirectory(), configFile[2:])
	}

	return
}

func getUserHomeDirectory() string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return dir
}
