package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/Sackbuoy/gameserver-operator/internal/crds"
	"github.com/Sackbuoy/gameserver-operator/internal/manager"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type Watcher struct {
	k8sClient *dynamic.DynamicClient
	logger    *zap.Logger
	watcher   watch.Interface
	manager   *manager.Manager
}

func New(ctx context.Context,
	logger *zap.Logger,
	k8sClient *dynamic.DynamicClient,
	gvc schema.GroupVersionResource,
	manager *manager.Manager,
) (*Watcher, error) {
	watcher, err := k8sClient.Resource(gvc).Namespace("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return &Watcher{
		logger:    logger,
		k8sClient: k8sClient,
		watcher:   watcher,
		manager:   manager,
	}, nil
}

func (w *Watcher) Watch() error {
	for event := range w.watcher.ResultChan() {
		obj, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		var crd crds.GameServer

		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &crd)
		if err != nil {
			w.logger.Error("Failed to convert CRD resource to Go struct",
				zap.String("Name: ", obj.GetName()),
				zap.Error(err))
		}

		// Get details from the game instance
		name := obj.GetName()
		namespace := obj.GetNamespace()
		key := fmt.Sprintf("%s/%s", namespace, name)
		creationTimestamp := obj.GetCreationTimestamp()

		// Only notify for new games
		switch event.Type {
		case "ADDED":
			if time.Since(creationTimestamp.Time).Seconds() > 60 {
				continue
			}

			w.logger.Info("CRD Event detected", zap.String("Key", key))
			w.logger.Info("With Event Type", zap.String("EventType", string(event.Type)))

			// Display some fields from the Game
			w.logger.Info("Found Game", zap.String("Name", name))

			err = w.manager.Create(obj.Object, crd.Spec.GameType, name, namespace)
			if err != nil {
				w.logger.Error("Error creating resources", zap.Error(err))

				continue
			}
		case "MODIFIED":
			w.logger.Info("CRD Event detected", zap.String("Key", key))
			w.logger.Info("With Event Type", zap.String("EventType", string(event.Type)))

			// Display some fields from the Game
			w.logger.Info("Found Game", zap.String("Name", name))

			err = w.manager.Update()
			if err != nil {
				w.logger.Error("Error updating resources", zap.Error(err))

				continue
			}
		case "DELETED":
			w.logger.Info("CRD Event detected", zap.String("Key", key))
			w.logger.Info("With Event Type", zap.String("EventType", string(event.Type)))

			// Display some fields from the Game
			w.logger.Info("Found Game", zap.String("Name", name))

			err = w.manager.Delete(name, namespace)
			if err != nil {
				w.logger.Error("Error deleting resources", zap.Error(err))

				continue
			}
		}
	}

	return nil
}
