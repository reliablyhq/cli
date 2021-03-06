package init

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
	crm "google.golang.org/api/cloudresourcemanager/v3"
	compute "google.golang.org/api/compute/v1"
)

var (
	gcpOptions          = []question.AskOpt{question.Subquestion}
	googleResourceTypes = []string{"Google Cloud Load Balancers"}
)

type gcpResource struct {
	ProjectID string
	//Region       string
	ResourceType string
	ResourceName string
}

func buildGCPResourceID() (r *gcpResource) {
	var projectID string
	var sanitizedResourceType string
	var resourceName string

	ctx := context.Background()
	crmService, err := crm.NewService(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			err = errors.New("you are not currently logged in to Google Cloud. Run `gcloud auth application-default login` to log in")
		}

		handleGCPError(err)
		return
	}

	orgsService := crm.NewOrganizationsService(crmService)
	orgs, err := orgsService.Search().Context(ctx).Do()

	if err != nil {
		handleGCPError(err)
		return
	}

	var orgsList = []string{}
	orgsMap := make(map[string]string)

	if orgs != nil && len(orgs.Organizations) > 0 {
		for _, o := range orgs.Organizations {
			displayName := o.DisplayName
			id := o.Name
			orgsList = append(orgsList, displayName)
			orgsMap[displayName] = id
		}
		orgDisplayName := question.WithSingleChoiceAnswer("Select an Organization.", gcpOptions, orgsList...)
		orgID := orgsMap[orgDisplayName]

		projectsService := crm.NewProjectsService(crmService)
		projects, err := projectsService.List().Context(ctx).Parent(orgID).Do()
		if err != nil {
			handleGCPError(err)
			return
		}
		if len(projects.Projects) == 0 {
			fmt.Println("  Reliably couldn't find any project. Check you have all the required permissions.")
			return
		}
		var projectsList = []string{}
		projectsMap := make(map[string]string)
		for _, p := range projects.Projects {
			displayName := p.DisplayName
			id := p.ProjectId
			fullName := displayName + " (" + id + ")"
			projectsList = append(projectsList, fullName)
			projectsMap[fullName] = id
		}
		// TODO Handle empty list case
		projectFullName := question.WithSingleChoiceAnswer("Select an Project.", gcpOptions, projectsList...)
		projectID = projectsMap[projectFullName]

		resourceType := question.WithSingleChoiceAnswer("What is the 'type' of the resource?", gcpOptions, googleResourceTypes...)
		sanitizedResourceType = sanitizeString(resourceType)

		lbctx := context.Background()
		computeService, _ := compute.NewService(lbctx)
		lbsService := compute.NewUrlMapsService(computeService)
		lbs, err := lbsService.List(projectID).Context(ctx).Do()
		if err != nil {
			handleGCPError(err)
			return
		}
		if len(lbs.Items) > 0 {
			var lbsList = []string{}
			for _, lb := range lbs.Items {
				name := lb.Name
				lbsList = append(lbsList, name)
			}
			resourceName = question.WithSingleChoiceAnswer("Select a resource.", gcpOptions, lbsList...)
		} else {
			fmt.Println(iconWarn, "Reliably couldn't find matching resources.")
			fmt.Println("  Cancelling.")
			return
		}

	} else {
		fmt.Println(iconWarn, "Reliably couldn't list your GCP Organizations.")
		fmt.Println("For interactive mode to work, you need to ensure the following conditions are met:")
		fmt.Println(" - You are currently logged in to Google Cloud. Run `gcloud auth application-default login` to log in.")
		fmt.Println(" - The currently-logged user has the `resourcemanager.projects.list` rights on the organization you're working on.")
		fmt.Println(" - The Cloud Resource Manager API is activated on the projects for which you want to list resources.")

		var insufficientGCPRightsOptions = []string{
			"Manually add resource (you will be asked to provide the project ID and resource name",
			"Cancel this resource",
		}
		insufficientGCPRights := question.WithSingleChoiceAnswer("What do you want to do now?", gcpOptions, insufficientGCPRightsOptions...)

		if insufficientGCPRights == "Cancel this resource" {
			return
		} else {
			projectID = question.WithStringAnswer("Enter the Project ID:", gcpOptions)
			resourceType := question.WithSingleChoiceAnswer("What is the 'type' of the resource?", gcpOptions, googleResourceTypes...)
			sanitizedResourceType = sanitizeString(resourceType)
			resourceName = question.WithStringAnswer("Enter the Resource name:", gcpOptions)
		}
	}

	//resourceID := fmt.Sprintf("%s/%s/%s", projectID, sanitizedResourceType, resourceName)

	r = &gcpResource{
		ProjectID:    projectID,
		ResourceType: sanitizedResourceType,
		ResourceName: resourceName,
	}

	return
}

func handleGCPError(err error) {
	errString := fmt.Sprintf("%+v\n", err)
	errStringSlice := strings.Split(errString, ".,")
	errStringNoSuffix := errStringSlice[0] + "."
	cleanErrString := strings.TrimPrefix(errStringNoSuffix, "googleapi: ")
	fmt.Println(color.Bold(color.Red(iconWarn, "GCP Error:")))
	fmt.Printf("%s\n", cleanErrString)
}
