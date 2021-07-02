package api

import "github.com/reliablyhq/cli/core/entities"

type QueryBody struct {
	Kind    string            `json:"kind"`
	Limit   int               `json:"limit"`
	Labels  map[string]string `json:"labels"`
	ForEach ForEach           `json:"forEach"`
}

type ForEach struct {
	ObjectiveResult ObjectiveResult `json:"objectiveResult"`
}

type ObjectiveResult struct {
	Include bool `json:"include"`
	Limit   int  `json:"limit"`
}

type QueryResponse struct {
	Objectives []ExpandedObjective `json:"objectives,omitempty"`
}

type ExpandedObjective struct {
	entities.Objective `json:"objective,omitempty"`
	ForEach            ForEachResponse `json:"forEach,omitempty"`
}

type ForEachResponse struct {
	ObjectiveResults []entities.ObjectiveResult `json:"objectiveResults,omitempty"`
}
