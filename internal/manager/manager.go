package manager

import (
	// "context"
	"fmt"
	"os"

	"github.com/Sackbuoy/gameserver-operator/internal/crds"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Manager struct {
	helmSettings          *cli.EnvSettings
	installedCharts       []*release.Release
  instanceMap *crds.CRDInstanceMap
	k8sClient             *dynamic.DynamicClient
	logger                *zap.Logger
	logOutput             func(string, ...any)
}

func New(k8sClient *dynamic.DynamicClient, logger *zap.Logger, instanceMap *crds.CRDInstanceMap) (*Manager, error) {
	settings := cli.New()

	logOutput := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logger.Info(msg)
	}
	// Create a new Helm install action
	return &Manager{
		helmSettings:          settings,
		k8sClient:             k8sClient,
		instanceMap: instanceMap,
		logger:                logger,
		logOutput:             logOutput,
	}, nil
}

func (m *Manager) Create(crdObject map[string]any, chartName, releaseName, namespace string) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(m.helmSettings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), m.logOutput); err != nil {
		m.logger.Error("Failed to initialize Helm Action Config", zap.Error(err))
	}

	installer := action.NewInstall(actionConfig)
	installer.Namespace = namespace
	installer.ReleaseName = releaseName

	gameServer, err := mapToGameServer(crdObject)
	if err != nil {
		m.logger.Fatal("Failed to install chart: %v", zap.Error(err))
	}

	chartPath := fmt.Sprintf("/charts/%s", chartName)

	chart, err := loader.Load(chartPath)
	if err != nil {
		m.logger.Fatal("Failed to load chart: %v", zap.Error(err))
	}

	valuesOverride := make(map[string]any)

	err = yaml.Unmarshal([]byte(gameServer.Spec.HelmChart.ValuesOverride), &valuesOverride)
	if err != nil {
		m.logger.Fatal("Failed to parse values Overrides")
	}

	chartInstall, err := installer.Run(chart, valuesOverride)
	if err != nil {
		m.logger.Fatal("Failed to install chart: %v", zap.Error(err))
	}

  err = m.instanceMap.Create(gameServer)
	if err != nil {
		m.logger.Fatal("Failed to add instance to internal cache", zap.Error(err))
	}

	m.logger.Info("Successfully installed release", zap.String("ReleaseName", chartInstall.Name))

	return nil
}

// TODO.
func (m *Manager) Read() error {
	return nil
}

// TODO.
func (m *Manager) Update() error {
	return nil
}

func (m *Manager) Delete(releaseName, namespace string) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(m.helmSettings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), m.logOutput); err != nil {
		m.logger.Error("Failed to initialize Helm Action Config", zap.Error(err))
	}

	installedCharts, err := getInstalledCharts(actionConfig)
	if err != nil {
		return err
	}

	if len(installedCharts) == 0 {
		return fmt.Errorf("No charts were found in %s", namespace)
	}

	// Fix: Check all charts before returning error
	found := false
	for _, chart := range installedCharts {
			if releaseName == chart.Name {
					found = true
					break
			}
	}

	if !found {
			return fmt.Errorf("%s not installed in namespace %s", releaseName, namespace)
	}

	uninstaller := action.NewUninstall(actionConfig)

	result, err := uninstaller.Run(releaseName)
	if err != nil {
		m.logger.Fatal("Failed run helm uninstall: %v", zap.Error(err))
	}

  err = m.instanceMap.Delete(releaseName)
	if err != nil {
		m.logger.Error("Failed to add instance to internal cache", zap.Error(err))
	}

	m.logger.Info(result.Info)

	return nil
}

