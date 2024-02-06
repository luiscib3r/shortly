package datasources

import (
	"os"
	"strconv"
)

type Environment struct {
	BaseURL string
	PORT    int
}

type EnvironmentDataSource struct {
	environment Environment
}

func NewEnvironmentDataSource() *EnvironmentDataSource {
	// Parse int from string
	var port int
	var err error

	if port, err = strconv.Atoi(os.Getenv("PORT")); port == 0 || err != nil {
		port = 8080
	}

	return &EnvironmentDataSource{
		environment: Environment{
			BaseURL: os.Getenv("BaseURL"),
			PORT:    port,
		},
	}
}

func (eds EnvironmentDataSource) GetEnvironment() *Environment {
	return &eds.environment
}
