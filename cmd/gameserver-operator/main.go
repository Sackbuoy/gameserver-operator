package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/Sackbuoy/gameserver-operator/internal/crds"
	"github.com/Sackbuoy/gameserver-operator/internal/manager"
	"github.com/Sackbuoy/gameserver-operator/internal/reconciler"
	"github.com/Sackbuoy/gameserver-operator/internal/watcher"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	ctx := context.Background()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// Try to use in-cluster config first, fall back to kubeconfig file
	config, err := rest.InClusterConfig()
	if err != nil {
		// Create config from kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			logger.Fatal("Error building kubeconfig", zap.Error(err))
		}
	}

	// Create a dynamic client for working with custom resources
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Fatal("Error creating dynamic client", zap.Error(err))
	}

	// Define the resource to watch - Game CRD
	gvc := schema.GroupVersionResource{
		Group:    "goopy.us",
		Version:  "v1",
		Resource: "gameservers",
	}

	logger.Info("Watching for GameServer instances...")

	// Get existing Game instances to avoid duplicate notifications
	existingGames, err := dynamicClient.Resource(gvc).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error("Error listing existing games", zap.Error(err))
	}

	existingGameMap := make(map[string]bool)

	for _, game := range existingGames.Items {
		key := fmt.Sprintf("%s/%s", game.GetNamespace(), game.GetName())
		existingGameMap[key] = true
	}

	var wg sync.WaitGroup
	instanceMap, err := crds.NewInstanceMap()
	if err != nil {
		logger.Error("Failed to make CRD instance map", zap.Error(err))
	}

	manager, err := manager.New(dynamicClient, logger, instanceMap)
	if err != nil {
		logger.Fatal("Error creating manager", zap.Error(err))
	}

	watcher, err := watcher.New(ctx, logger, dynamicClient, gvc, manager)
	if err != nil {
		logger.Fatal("Error creating watcher", zap.Error(err))
	}

	reconciler, err := reconciler.New(ctx, logger, manager, dynamicClient, gvc, instanceMap)
	if err != nil {
		logger.Fatal("Error creating reconciler", zap.Error(err))
	}

	// Watch for events
	wg.Add(1)
	go func() {
		defer wg.Done()
		watcher.Watch()
	}()

	// periodically loop through CRD instances to ensure corresponding resources
	// are synced
	wg.Add(1)
	go func() {
		defer wg.Done()
		reconciler.MonitorLoop(ctx, "")
	}()

	wg.Wait()
}
