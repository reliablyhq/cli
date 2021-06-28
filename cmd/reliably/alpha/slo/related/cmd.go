package related

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/manifest"
	"github.com/reliablyhq/cli/embedded/nodegraph"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Options struct {
	IO           *iostreams.IOStreams
	ManifestPath string
	Raw          bool
	Port         string
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
	$ reliably alpha slo related

	%s
	$ reliably alpha slo related -m reliably.yaml --port 8085

	%s
	$ reliably alpha slo related --raw`,
	color.Grey("open visualisation on a random port between 60000-61000"),
	color.Grey("open visualisation app on port 8085"),
	color.Grey("return raw JSON blob of visualisation data"),
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
			org, err := api.CurrentUserOrganization(client, config.Hostname)
			if err != nil {
				return err
			}

			if opts.Raw {
				g, err := api.GetRelationshipGraph(client, config.EntityServerHost, org.Name, m)
				if err != nil {
					return err
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(g)
			}

			return serveRelationshipGraph(client, org.Name, opts.Port, opts.ManifestPath)
		},
	}

	// define flags
	cmd.Flags().StringVarP(&opts.ManifestPath, "manifest", "m", manifest.DefaultManifestPath, "the location of the manifest file")
	cmd.Flags().StringVarP(&opts.Port, "port", "p", "", "the port to serve the graph visualisation on. A random port [60000-61000] is used if no port is profided")
	cmd.Flags().BoolVarP(&opts.Raw, "raw", "r", false, "prints raw json graph data")
	return cmd
}

func serveRelationshipGraph(client *api.Client, org, port, manifestPath string) error {
	logger := log.StandardLogger()
	if port == "" {
		port = fmt.Sprintf("%d", randomInt(60000, 61000))
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
			logger.Debugf("error loading manifest file: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		g, err := api.GetRelationshipGraph(client, config.EntityServerHost, org, m)
		if err != nil {
			logger.Debugf("error fetching relationship data from API: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(g); err != nil {
			logger.Debugf("error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	openbrowser(uri)
	fmt.Println(color.Green("serving relationship graph on:"), color.Cyan(uri))
	fmt.Println(color.Green("opening browser..."))
	return http.ListenAndServe(port, server)
}

// Returns an int >= min, < max
func randomInt(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return min + int64(rand.Int63n(max-min))
}

func openbrowser(url string) {
	var err error
	fmt.Println()
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
