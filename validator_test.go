package jsonunknown

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

// Configuration for kubernetes prow tool (modified version) is used as example
// source: https://github.com/kubernetes/test-infra/blob/3853dfce6b8c182809f9df85c4d9271f705fabf0/prow/plugins/config.go
type Configuration struct {
	Plugins       map[string][]string `json:"plugins,omitempty"`
	ConfigUpdater ConfigUpdater       `json:"config_updater,omitempty"`
	Size          *Size               `json:"size,omitempty"`
	Triggers      []Trigger           `json:"triggers,omitempty"`
	Version       int                 `json:"version"`
}

// Trigger defines trigger config
type Trigger struct {
	Repos []string `json:"repos,omitempty"`
}

// Size specifies configuration for the size plugin, defining lower bounds (in # lines changed) for each size label.
// XS is assumed to be zero.
type Size struct {
	S   int `json:"s"`
	M   int `json:"m"`
	L   int `json:"l"`
	Xl  int `json:"xl"`
	Xxl int `json:"xxl"`
}

// ConfigUpdater contains the configuration for the config-updater plugin.
type ConfigUpdater struct {
	Maps       map[string]ConfigMapSpec `json:"maps,omitempty"`
	ConfigFile string                   `json:"config_file,omitempty"`
	PluginFile string                   `json:"plugin_file,omitempty"`
}

// ConfigMapSpec contains configuration options for the configMap being updated
// by the config-updater plugin.
type ConfigMapSpec struct {
	Name       string   `json:"name"`
	Key        string   `json:"key,omitempty"`
	Namespace  string   `json:"namespace,omitempty"`
	Namespaces []string // json tag intentionally not provided
}

func TestInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)

	cfg := &Configuration{}
	bytes := []byte("definitely not json")
	_, err := ValidateUnknownFields(bytes, cfg)
	g.Expect(err).To(Not(BeNil()))
}

func TestUseCases(t *testing.T) {
	testCases := []struct {
		name        string
		configBytes []byte
		config      interface{}
		expected    []string
	}{
		{
			name: "valid config",
			configBytes: []byte(`plugins:
  kube/kube:
  - size
  - config-updater
config_updater:
  maps:
    # Update the plugins configmap whenever plugins.yaml changes
    kube/plugins.yaml:
      name: plugins
size:
  s: 1`),
			expected: nil,
		},
		{
			name: "invalid top-level struct property",
			configBytes: []byte(`plugins:
  kube/kube:
  - size
  - config-updater
notconfig_updater:
  maps:
    # Update the plugins configmap whenever plugins.yaml changes
    kube/plugins.yaml:
      name: plugins
size:
  s: 1`),
			expected: []string{"notconfig_updater"},
		},
		{
			name: "invalid top-level simple property",
			configBytes: []byte(`plugins:
  kube/kube:
  - size
  - config-updater
verzion: 1
size:
  s: 1`),
			expected: []string{"verzion"},
		},
		{
			name: "invalid second-level property",
			configBytes: []byte(`plugins:
  kube/kube:
  - size
  - config-updater
size:
  xs: 1
  s: 5`),
			expected: []string{"size.xs"},
		},
		{
			name: "invalid top-level array element",
			configBytes: []byte(`plugins:
  kube/kube:
  - size
  - trigger
triggers:
- repos:
  - kube/kube
- repoz:
  - kube/kubez`),
			expected: []string{"triggers[1].repoz"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			cfg := &Configuration{}
			if err := yaml.Unmarshal(tc.configBytes, cfg); err != nil {
				t.Fatalf("Unable to unmarhsal yaml: %v", err)
			}
			jsonCfgBytes, err := yaml.YAMLToJSON(tc.configBytes)
			g.Expect(err).To(BeNil())
			got, err := ValidateUnknownFields(jsonCfgBytes, cfg)
			g.Expect(err).To(BeNil())
			g.Expect(got).To(HaveLen(len(tc.expected)))
			for _, elem := range tc.expected {
				g.Expect(got).To(ContainElement(elem))
			}
		})
	}
}
