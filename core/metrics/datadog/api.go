package datadog

import (
	"context"
	"encoding/json"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	log "github.com/sirupsen/logrus"
)

func ValidateApiKey() (bool, error) {

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
