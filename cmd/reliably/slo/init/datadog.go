package init

import (
	"context"
	"encoding/json"

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
