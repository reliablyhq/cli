package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/utils"
	log "github.com/sirupsen/logrus"
)

func CreateEntity(client *Client, hostname string, org string, entity entities.Entity) error {

	var version string
	version = entity.Version()
	version = strings.ToLower(version)

	var kind string
	kind = plural(entity.Kind())
	kind = strings.ToLower(kind)

	shortVersion, ok := utils.GetShortVersion(version)
	if !ok {
		return fmt.Errorf("version %v not supported", version)
	}

	path := requestPath(shortVersion, kind, org)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(entity); err != nil {
		return err
	}

	return client.RESTv2(hostname, http.MethodPut, path, &body, nil)
}

func GetObjectiveResults(client *Client, hostname string, version string, org string) (*[]entities.ObjectiveResultResponse, error) {

	var entitiesResult *[]entities.ObjectiveResultResponse

	shortVersion, ok := utils.GetShortVersion(version)
	if !ok {
		return nil, fmt.Errorf("version %v not supported", version)
	}

	path := requestPath(shortVersion, "objective-results", org)
	err := client.RESTv2(hostname, http.MethodGet, path, nil, &entitiesResult)

	if err != nil {
		return nil, fmt.Errorf("failed to make API call: %w", err)
	}

	return entitiesResult, nil

}

func GetRelationshipGraph(client *Client, hostname, org string, m entities.Manifest) (*entities.NodeGraph, error) {
	if len(m) == 0 {
		return nil, errors.New("no entities found in manifest")
	}

	// TODO: by using m[0].Version() we assumes all entities in a manifest
	// will have the same API version. This should be changed if/when the API
	// is extended beyond v1
	path := fmt.Sprintf("%s/%s/%s/objectives/relatedto", "entities", m[0].Version(), org)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(m); err != nil {
		return nil, err
	}

	var g entities.NodeGraph
	return &g, client.RESTv2(hostname,
		http.MethodPost, path, &body, &g)
}

// SyncManifest - synchronize objectives from manifest with the entity server
func SyncManifest(client *Client, entityHost, org string, m entities.Manifest) error {
	var hasErr bool
	var errchan = make(chan error, len(m))
	var w sync.WaitGroup

	for _, slo := range m {
		w.Add(1)
		go func(slo *entities.Objective) {
			defer w.Done()
			log.Debugf("syncing slo: %s", slo.Name)
			if err := CreateEntity(client, entityHost, org, slo); err != nil {
				errchan <- fmt.Errorf("error syncing manifest object: %s - %s", slo.Name, err)
			}
		}(slo)
	}

	w.Wait()
	close(errchan)

	for e := range errchan {
		log.Debug(e)
		hasErr = true
	}

	if hasErr {
		return errors.New("An error occured while syncing your manifest")
	}

	return nil
}

func requestPath(version, kind, org string) string {

	return fmt.Sprintf("%s/%s/%s/%s", "entities", version, org, kind)
}

// plural returns the puralized string,
// append trailing 's' if not already ending with it
func plural(s string) string {
	if !strings.HasSuffix(s, "s") {
		s = fmt.Sprintf("%ss", s)
	}
	return s
}
