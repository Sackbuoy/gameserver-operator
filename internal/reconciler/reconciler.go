package reconciler

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/Sackbuoy/gameserver-operator/internal/crds"
	"github.com/Sackbuoy/gameserver-operator/internal/util"
	"github.com/Sackbuoy/gameserver-operator/internal/manager"
)

type Reconciler struct {
	helmSettings *cli.EnvSettings
	logger       *zap.Logger
	k8sClient    *dynamic.DynamicClient
	loopInterval time.Duration
	instanceMap  *crds.CRDInstanceMap
	manager      *manager.Manager
	logOutput    func(string, ...any)
  gvc schema.GroupVersionResource
}

func New(ctx context.Context, logger *zap.Logger, manager *manager.Manager, k8sClient *dynamic.DynamicClient, gvc schema.GroupVersionResource, instanceMap *crds.CRDInstanceMap) (*Reconciler, error) {
	settings := cli.New()

	logOutput := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logger.Info(msg)
	}

	return &Reconciler{
		logger:       logger,
		loopInterval: time.Second * 5,
		manager:      manager,
		helmSettings: settings,
		logOutput:    logOutput,
    gvc: gvc,
    instanceMap: instanceMap,
    k8sClient: k8sClient,
	}, nil
}

func (r *Reconciler) MonitorLoop(ctx context.Context, namespace string) {
	ticker := time.NewTicker(r.loopInterval)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(r.helmSettings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), r.logOutput); err != nil {
		r.logger.Error("Failed to initialize Helm Action Config", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
      err := r.reconcile(ctx, *actionConfig)
      if err != nil {
        r.logger.Error("Error listing existing games", zap.Error(err))
      }
			// loop through currently tracked CRDs, check if the helm chart is installed
		}
	}
}

func (r *Reconciler) reconcile(ctx context.Context, actionConfig action.Configuration) error {
	unstructuredList, err := r.k8sClient.Resource(r.gvc).Namespace("").List(ctx, metav1.ListOptions{})
  if err != nil {
    return err
  }

  client := action.NewGet(&actionConfig)
	
	for _, instance := range unstructuredList.Items {
		var newCRD crds.GameServer
    releaseName := instance.GetName()

		util.MapUnstructuredToStruct(&instance, &newCRD)
		err = r.instanceMap.Create(&newCRD)
    if err != nil {
      return err
    }

    // Try to get the release
    release, err := client.Run(releaseName)
    switch release {
    case nil:
      // if release is nil and err isn't helm chart is not installed
      if err != nil {
        r.logger.Info("No Helm chart release found. Installing...", zap.String("Instance", releaseName))
        r.manager.Create(instance.Object, newCRD.Spec.GameType, releaseName, instance.GetNamespace())
      } else {
        // this would be weird, should never happen
        r.logger.Error("No helm chart release found, but no error was returned", zap.String("Instance", releaseName))
      }
    default:
      if err != nil {
        // this would be weird, should never happen
        r.logger.Error("Helm chart release found, but error was returned", zap.String("Instance", releaseName), zap.Error(err))
      } else {
        // chart exists, no action needed
        continue
      }
    }
	}

  return nil
}
