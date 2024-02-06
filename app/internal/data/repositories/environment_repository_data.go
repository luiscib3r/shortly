package repositories

import (
	"github.com/luiscib3r/shortly/app/internal/data/datasources"
)

type EnvironmentRepositoryData struct {
	environment *datasources.EnvironmentDataSource
}

func NewEnvironmentRepositoryData(
	environment *datasources.EnvironmentDataSource,
) *EnvironmentRepositoryData {
	return &EnvironmentRepositoryData{
		environment: environment,
	}
}

func (r EnvironmentRepositoryData) GetBaseUrl() string {

	return r.environment.GetEnvironment().BaseURL
}
