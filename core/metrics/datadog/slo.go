package datadog

import (
	"context"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

func ListDatadogSLOs() (datadog.SLOListResponse, error) {

	ctx := datadog.NewDefaultContext(context.Background())

	optionalParams := datadog.ListSLOsOptionalParameters{}

	configuration := datadog.NewConfiguration()

	apiClient := datadog.NewAPIClient(configuration)
	resp, _, err := apiClient.ServiceLevelObjectivesApi.ListSLOs(ctx, optionalParams)
	return resp, err
}

func GetSLOHistory(sloId string, from time.Time, to time.Time, target float64) (datadog.SLOHistoryResponse, error) {

	ctx := datadog.NewDefaultContext(context.Background())

	configuration := datadog.NewConfiguration()
	configuration.SetUnstableOperationEnabled("GetSLOHistory", true)

	// Using target is fixed now https://github.com/DataDog/datadog-api-client-go/issues/951
	// still need to be tested ...
	/*
		optionalParams := datadog.GetSLOHistoryOptionalParameters{
			Target: &target,
		}
	*/

	apiClient := datadog.NewAPIClient(configuration)
	resp, _, err := apiClient.ServiceLevelObjectivesApi.GetSLOHistory(
		ctx, sloId, from.UTC().Unix(), to.UTC().Unix()) //, optionalParams)

	return resp, err
}
