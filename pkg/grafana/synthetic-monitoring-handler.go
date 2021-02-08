package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grizzly/pkg/grizzly"
	"github.com/grafana/tanka/pkg/kubernetes/manifest"
	"github.com/mitchellh/mapstructure"
)

/*
 * @TODO
 * 1. The API does not have a GET method, so we have to fake it here
 * 2. The API expects an ID and a tenantId in an update, but these are
 *    generated by the server so cannot be represented in Jsonnet.
 *    Therefore, we have to pre-retrieve the check to get those values
 *    so we can inject them before posting JSON.
 * 3. This means pre-retrieving the check *twice*, once to establish
 *    whether this resource has changed or not (within Grizzly ifself)
 *    and again within this provider to retrieve IDs. Not ideal.
 */

// SyntheticMonitoringHandler is a Grizzly Provider for Grafana Synthetic Monitoring
type SyntheticMonitoringHandler struct{}

// NewSyntheticMonitoringHandler returns configuration defining a new Grafana Provider
func NewSyntheticMonitoringHandler() *SyntheticMonitoringHandler {
	return &SyntheticMonitoringHandler{}
}

// GetName returns the name for this provider
func (h *SyntheticMonitoringHandler) GetName() string {
	return "synthetic-monitor"
}

// GetFullName returns the name for this provider
func (h *SyntheticMonitoringHandler) GetFullName() string {
	return "grafana.synthetic-monitor"
}

const syntheticMonitoringChecksPath = "syntheticMonitoring"

// GetJSONPaths returns paths within Jsonnet output that this provider will consume
func (h *SyntheticMonitoringHandler) GetJSONPaths() []string {
	return []string{
		syntheticMonitoringChecksPath,
	}
}

// GetExtension returns the file name extension for a check
func (h *SyntheticMonitoringHandler) GetExtension() string {
	return "json"
}

// APIVersion returns the api version for this resource
func (h *SyntheticMonitoringHandler) APIVersion() string {
	return "grafana.com/grizzly/v1"
}

// Kind returns the resource kind for this type of resource
func (h *SyntheticMonitoringHandler) Kind() string {
	return "SyntheticMonitoringCheck"
}
func (h *SyntheticMonitoringHandler) newCheckResource(filename string, check Check) grizzly.Resource {
	resource := grizzly.Resource{
		UID:      check.UID(),
		Filename: filename,
		Handler:  h,
		Detail:   check,
		JSONPath: syntheticMonitoringChecksPath,
	}
	return resource
}

// ParseHiddenElements parses an interface{} object into a struct for this resource type
func (h *SyntheticMonitoringHandler) ParseHiddenElements(path string, i interface{}) (grizzly.ResourceList, error) {
	resources := grizzly.ResourceList{}
	msi := i.(map[string]interface{})
	for k, v := range msi {
		m := grizzly.NewManifest(h.APIVersion(), h.Kind(), k, v)
		resource, err := h.Parse(m)
		if err != nil {
			return nil, err
		}
		resources[resource.Key()] = *resource
	}
	return resources, nil
}

// Parse parses a single resource from an interface{} object
func (h *SyntheticMonitoringHandler) Parse(m manifest.Manifest) (*grizzly.Resource, error) {
	check := Check{}
	err := mapstructure.Decode(m["spec"], &check)
	if err != nil {
		return nil, err
	}
	resource := h.newCheckResource(m.Metadata().Name(), check)
	return &resource, nil
}

// Unprepare removes unnecessary elements from a remote resource ready for presentation/comparison
func (h *SyntheticMonitoringHandler) Unprepare(resource grizzly.Resource) *grizzly.Resource {
	delete(resource.Detail.(Check), "tenantId")
	delete(resource.Detail.(Check), "id")
	delete(resource.Detail.(Check), "modified")
	delete(resource.Detail.(Check), "created")
	return &resource
}

// Prepare gets a resource ready for dispatch to the remote endpoint
func (h *SyntheticMonitoringHandler) Prepare(existing, resource grizzly.Resource) *grizzly.Resource {
	resource.Detail.(Check)["tenantId"] = existing.Detail.(Check)["tenantId"]
	resource.Detail.(Check)["id"] = existing.Detail.(Check)["id"]
	return &resource
}

// GetByUID retrieves JSON for a resource from an endpoint, by UID
func (h *SyntheticMonitoringHandler) GetByUID(UID string) (*grizzly.Resource, error) {
	check, err := getRemoteCheck(UID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving check %s: %v", UID, err)
	}
	resource := h.newCheckResource("", *check)
	return &resource, nil
}

// GetRepresentation renders a resource as JSON or YAML as appropriate
func (h *SyntheticMonitoringHandler) GetRepresentation(uid string, resource grizzly.Resource) (string, error) {
	j, err := json.MarshalIndent(resource.Detail, "", "  ")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// GetRemoteRepresentation retrieves a datasource as JSON
func (h *SyntheticMonitoringHandler) GetRemoteRepresentation(uid string) (string, error) {
	check, err := getRemoteCheck(uid)
	if err != nil {
		return "", err
	}
	return check.toJSON()
}

// GetRemote retrieves a datasource as a Resource
func (h *SyntheticMonitoringHandler) GetRemote(uid string) (*grizzly.Resource, error) {
	check, err := getRemoteCheck(uid)
	if err != nil {
		return nil, err
	}
	resource := h.newCheckResource("", *check)
	return &resource, nil
}

// Add adds a new check to the SyntheticMonitoring endpoint
func (h *SyntheticMonitoringHandler) Add(resource grizzly.Resource) error {
	url := getURL("api/v1/check/add")
	return postCheck(url, newCheck(resource))
}

// Update pushes an updated check to the SyntheticMonitoring endpoing
func (h *SyntheticMonitoringHandler) Update(existing, resource grizzly.Resource) error {
	check := newCheck(resource)
	url := getURL("api/v1/check/update")
	return postCheck(url, check)
}

// Preview renders Jsonnet then pushes them to the endpoint if previews are possible
func (h *SyntheticMonitoringHandler) Preview(resource grizzly.Resource, notifier grizzly.Notifier, opts *grizzly.PreviewOpts) error {
	return grizzly.ErrNotImplemented
}
