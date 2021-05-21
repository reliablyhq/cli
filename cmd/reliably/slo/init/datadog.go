package init

import (
	"context"
	"encoding/json"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	log "github.com/sirupsen/logrus"
)

func validateApiKey() (bool, error) {

	ctx := datadog.NewDefaultContext(context.Background())

	configuration := datadog.NewConfiguration()

	apiClient := datadog.NewAPIClient(configuration)

	resp, r, err := apiClient.AuthenticationApi.Validate(ctx)
	if err != nil {
		log.Debugf("Error when calling `AuthenticationApi.Validate`: %v\n", err)
		log.Debugf("Full HTTP response: %v\n", r)
		return false, err
	}
	// response from `Validate`: AuthenticationValidationResponse
	responseContent, _ := json.MarshalIndent(resp, "", "  ")
	log.Debugf("Response from AuthenticationApi.Validate:\n%s\n", responseContent)

	return *resp.Valid, nil

}

func ListDatadogSLOs() (datadog.SLOListResponse, error) {

	ctx := datadog.NewDefaultContext(context.Background())

	optionalParams := datadog.ListSLOsOptionalParameters{}

	configuration := datadog.NewConfiguration()

	apiClient := datadog.NewAPIClient(configuration)
	resp, _, err := apiClient.ServiceLevelObjectivesApi.ListSLOs(ctx, optionalParams)
	return resp, err
}

/*

func main() {
    ctx := datadog.NewDefaultContext(context.Background())

    ids := "id1, id2, id3" // string | A comma separated list of the IDs of the service level objectives objects. (optional)
    query := "monitor" // string | The query string to filter results based on SLO names. (optional)
    tagsQuery := "env:prod" // string | The query string to filter results based on a single SLO tag. (optional)
    metricsQuery := "aws.elb.request_count" // string | The query string to filter results based on SLO numerator and denominator. (optional)
    optionalParams := datadog.ListSLOsOptionalParameters{
        Ids: &ids,
        Query: &query,
        TagsQuery: &tagsQuery,
        MetricsQuery: &metricsQuery,
    }

    configuration := datadog.NewConfiguration()

    apiClient := datadog.NewAPIClient(configuration)
    resp, r, err := apiClient.ServiceLevelObjectivesApi.ListSLOs(ctx, optionalParams)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ServiceLevelObjectivesApi.ListSLOs`: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListSLOs`: SLOListResponse
    responseContent, _ := json.MarshalIndent(resp, "", "  ")
    fmt.Fprintf(os.Stdout, "Response from ServiceLevelObjectivesApi.ListSLOs:\n%s\n", responseContent)
}
*/

func GetSLOHistory(sloId string, from time.Time, to time.Time, target float64) (datadog.SLOHistoryResponse, error) {

	/*
			*
			* Note, when we inspect network requests in browser,
			* datadog UI is using a timeframe query param that is not available on the go client
			* https://app.datadoghq.eu/api/v1/slo/04e0719b828a548a8d34cadd6bc5b84a/history?interval=86400&group_uptime=false&no_cache=false&timeframe=30d
			* When timeframe is given, error budget is returned in the response
			*  "overall": {
		            "errors": null,
		            "sli_value": 100,
		            "precision": {
		                "7d": 1,
		                "30d": 1
		            },
		            "error_budget_remaining": {
		                "30d": 100
		            },
		            "corrections": [],
		            "span_precision": 1
		        },
	*/

	ctx := datadog.NewDefaultContext(context.Background())

	configuration := datadog.NewConfiguration()
	configuration.SetUnstableOperationEnabled("GetSLOHistory", true)

	// Investigate, when using target as optional param,
	// we get an error from api: custom is not a valid SLOTimeframe
	// raised an issue on DD : https://github.com/DataDog/datadog-api-client-go/issues/951
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

/*
func main() {
    ctx := datadog.NewDefaultContext(context.Background())

    sloId := "sloId_example" // string | The ID of the service level objective object.
    fromTs := int64(789) // int64 | The `from` timestamp for the query window in epoch seconds.
    toTs := int64(789) // int64 | The `to` timestamp for the query window in epoch seconds.
    target := float64(1.2) // float64 | The SLO target. If `target` is passed in, the response will include the error budget that remains. (optional)
    optionalParams := datadog.GetSLOHistoryOptionalParameters{
        Target: &target,
    }

    configuration := datadog.NewConfiguration()
    configuration.SetUnstableOperationEnabled("GetSLOHistory", true)

    apiClient := datadog.NewAPIClient(configuration)
    resp, r, err := apiClient.ServiceLevelObjectivesApi.GetSLOHistory(ctx, sloId, fromTs, toTs, optionalParams)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ServiceLevelObjectivesApi.GetSLOHistory`: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSLOHistory`: SLOHistoryResponse
    responseContent, _ := json.MarshalIndent(resp, "", "  ")
    fmt.Fprintf(os.Stdout, "Response from ServiceLevelObjectivesApi.GetSLOHistory:\n%s\n", responseContent)
}
*/
