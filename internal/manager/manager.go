package manager

import (
  "fmt"
  "os"
  "log"
  "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

type Manager struct {
}

func New() (*Manager, error) {
  return &Manager{}, nil
}

func (m *Manager) Create(url, name, namespace string) error {
  settings := cli.New()
	
	// Create a new Helm install action
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Fatalf("Failed to initialize action configuration: %v", err)
	}

	client := action.NewInstall(actionConfig)
	client.Namespace = namespace
	client.ReleaseName = name
	
	// Specify chart location (local or remote)
	// chartPath := "https://helm.github.io/examples" // or URL for remote charts
	chartPath := url // or URL for remote charts
	chart, err := loader.Load(chartPath)
	if err != nil {
		log.Fatalf("Failed to load chart: %v", err)
	}
	
	// Install the chart
	rel, err := client.Run(chart, nil) // second parameter is values
	if err != nil {
		log.Fatalf("Failed to install chart: %v", err)
	}
	
	fmt.Printf("Successfully installed release %s\n", rel.Name)
  return nil
}

// TODO
func (m *Manager) Read() error {
  return nil
}

// TODO
func (m *Manager) Update() error {
  return nil
}

func (m *Manager) Delete(name, namespace string) error {
  settings := cli.New()
	
	// Create a new Helm install action
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Fatalf("Failed to initialize action configuration: %v", err)
	}

  uninstaller := action.NewUninstall(actionConfig)
  result, err := uninstaller.Run(name)
  if err != nil {
		log.Fatalf("Failed run helm uninstall: %v", err)
  }

  fmt.Printf("%s\n", result.Info)
  
  return nil
}
