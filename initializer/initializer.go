package initializer

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/cldcvr/audit-db/config"
	"github.com/cldcvr/audit-db/internal/logger"
	"github.com/cldcvr/audit-db/pkg/connection"
	"github.com/cldcvr/audit-db/router"
)

// Init project
func Init() (*gin.Engine, chan bool) {
	logger.Init()

	if err := config.InitAll(); err != nil {
		log.Fatal().Msg(err.Error())
	}

	stop, err := createConsumersFromConfig()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Register consumers from config")

	return router.Init(), stop
}

func createConsumersFromConfig() (chan bool, error) {
	sub, _ := config.GetSubscription()
	gcp, _ := config.GCP()

	// create connection
	conn, err := connection.New(gcp.ProjectID)
	if err != nil {
		return nil, err
	}

	// Register handlers
	for _, table := range sub.Tables {
		if ok := validateConfig(table); !ok {
			return nil, errors.New("validationException: could not validate config")
		}
		if err := conn.RegisterHandler(table); err != nil {
			conn.ShutDown()
			return nil, err
		}
	}

	// start consumer
	if err := conn.StartConsumer(); err != nil {
		return nil, err
	}

	var stop = make(chan bool)
	go func(close chan bool) {
		<-stop
		conn.ShutDown()
	}(stop)
	return stop, nil
}

func validateConfig(tablename string) bool {
	// check if configuration is present for correct tables
	tables, _ := config.Tables()
	if _, ok := tables[tablename]; !ok {
		return false
	}
	return true
}
