package related

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/embedded/nodegraph"
	"github.com/reliablyhq/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Options struct {
	IO           *iostreams.IOStreams
	ManifestPath string
	Raw          bool
	Port         string
	Filters      []string
}

type OptFunc func(*Options) error

var longDesc = heredoc.Doc(`
	By defining the metadata.relatedTo keyword, arbitrary relationships
	between 2 or more objectives/entities can be described. 
	
	The [related] command uses the defined relationships to generate a 
	Node Graph visualisation of the relationships defined in the local 
	manifest and all entities within the organisation.
	
	Passing the --raw/-r flag will return the raw JSON data used to draw
	the graph. 
	
	NOTE: the raw JSON data can also be retrieved by going to /data
	when running the visualisation in a browser`)

var examples = heredoc.Docf(`
	%s
	$ reliably slo related

	%s
	$ reliably slo related -m reliably.yaml --port 8085

	%s
	$ reliably slo related --raw
	
	%s
	$ reliably slo related --filter 'key=value' --filter 'key=value' -m reliably.yaml`,
	color.Grey("open visualisation on a random port between 60000-61000"),
	color.Grey("open visualisation app on port 8085"),
	color.Grey("return raw JSON blob of visualisation data"),
	color.Grey("open visualisation app on random port, only showing nodes with labels matching the given filters"),
)

func NewCommand(runF OptFunc) *cobra.Command {
	opts := &Options{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:     "related",
		Short:   "fetches a node graph of relationships based on manifest objectives",
		Long:    longDesc,
		Example: examples,
		RunE: func(cmd *cobra.Command, args []string) error {
			// read manifest
			var m entities.Manifest
			if err := m.LoadFromFile(opts.ManifestPath); err != nil {
				return err
			}

			// get API client
			client := api.NewClientFromHTTP(api.AuthHTTPClient(config.Hostname))
			org, err := config.GetCurrentOrgInfo()
			if err != nil {
				return err
			}

			if opts.Raw {
				g, err := api.GetRelationshipGraph(client, config.EntityServerHost, org.Name, m)
				if err != nil {
					return err
				}

				g = applyFilters(g, opts.Filters...)
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(g)
			}

			return serveRelationshipGraph(client, org.Name, opts.Port, opts.ManifestPath, opts.Filters...)
		},
	}

	// define flags
	cmd.Flags().StringVarP(&opts.ManifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")
	cmd.Flags().StringVarP(&opts.Port, "port", "p", "", "the port to serve the graph visualisation on. A random port [60000-61000] is used if no port is profided")
	cmd.Flags().BoolVarP(&opts.Raw, "raw", "r", false, "prints raw json graph data")
	cmd.Flags().StringArrayVarP(&opts.Filters, "filters", "f", []string{}, "<key=value> labels to filter relationship graph nodes")
	return cmd
}

func serveRelationshipGraph(client *api.Client, org, port, manifestPath string, filters ...string) error {
	// used to hash manifest file
	var manifestHash string

	if port == "" {
		port = fmt.Sprintf("%d", utils.RandomInt(60000, 61000))
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	uri := fmt.Sprintf("http://localhost%s", port)

	// define server handlers
	rootfs, err := fs.Sub(nodegraph.FS, nodegraph.RootDir)
	if err != nil {
		return err
	}
	fs := http.FileServer(http.FS(rootfs))
	server := http.NewServeMux()
	server.Handle("/", fs)
	server.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		// read manifest
		var m entities.Manifest
		if err := m.LoadFromFile(manifestPath); err != nil {
			log.Debugf("error loading manifest file: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// resync manifest
		if manifestHash != m.Hash() {
			if err := api.SyncManifest(client, config.EntityServerHost, org, m); err != nil {
				log.Debugf("error syncing manifest: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			manifestHash = m.Hash()
		}

		g, err := api.GetRelationshipGraph(client, config.EntityServerHost, org, m)
		if err != nil {
			log.Debugf("error fetching relationship data from API: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		g = applyFilters(g, filters...)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(g); err != nil {
			log.Debugf("error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	utils.OpenInBrowser(uri)
	fmt.Println(color.Green("serving relationship graph on:"), color.Cyan(uri))
	fmt.Println(color.Green("opening browser..."))
	return http.ListenAndServe(port, server)
}

// used to hash manifest file
var manifestHash string

// sync manifest if there is a change
func syncManifest(client *api.Client, org string, m entities.Manifest) error {

	// return if manifestHash is the same
	if manifestHash == m.Hash() {
		return nil
	}

	defer func() {
		manifestHash = m.Hash()
	}()

	for _, slo := range m {
		log.Debugf("syncing: %s", slo.Name)
		if err := api.CreateEntity(client, config.EntityServerHost, org, slo); err != nil {
			return fmt.Errorf("error syncing manifest object: %s - %s", slo.Name, err)
		}
	}
	return nil
}

// applyFilters - filter graph based on user provided labels
func applyFilters(g *entities.NodeGraph, filters ...string) *entities.NodeGraph {
	if len(filters) == 0 {
		return g
	}

	var updatedGraph entities.NodeGraph
	filteredNodeIDs := make(map[string]struct{})
	validFilterReg := regexp.MustCompile(`^[a-zA-Z0-9\-\_]+=[a-zA-Z0-9\s\_\-]+$`)
	for _, f := range filters {
		if !validFilterReg.MatchString(f) {
			log.Debugf("error: invalid key=value pair detected for filter: [%s]", f)
			continue
		}

		fk, fv := strings.Split(f, "=")[0], strings.Split(f, "=")[1]
		for _, node := range g.Nodes {
			for mk, mv := range node.Metadata.Labels {
				if fk == mk && fv == mv {
					updatedGraph.Nodes = append(updatedGraph.Nodes, node)
					filteredNodeIDs[node.ID] = struct{}{}
				}
			}
		}

		// check both edge IDs exist in updated Graph
		for _, edge := range g.Edges {
			if _, ok := filteredNodeIDs[edge.Source]; !ok {
				continue
			}

			if _, ok := filteredNodeIDs[edge.Target]; !ok {
				continue
			}

			updatedGraph.Edges = append(updatedGraph.Edges, edge)
		}

	}
	return &updatedGraph
}
