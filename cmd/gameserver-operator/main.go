package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
  "time"

	"github.com/Sackbuoy/gameserver-operator/internal/manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
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
			fmt.Printf("Error building kubeconfig: %v\n", err)
			os.Exit(1)
		}
	}

	// Create a dynamic client for working with custom resources
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating dynamic client: %v\n", err)
		os.Exit(1)
	}

	// Define the resource to watch - Game CRD
	gamesGVR := schema.GroupVersionResource{
		Group:    "goopy.us",
		Version:  "v1",
		Resource: "gameservers",
	}

	fmt.Println("Watching for GameServer instances...")

	// Get existing Game instances to avoid duplicate notifications
	existingGames, err := dynamicClient.Resource(gamesGVR).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing existing games: %v\n", err)
		os.Exit(1)
	}

	existingGameMap := make(map[string]bool)
	for _, game := range existingGames.Items {
		key := fmt.Sprintf("%s/%s", game.GetNamespace(), game.GetName())
		existingGameMap[key] = true
	}

	// Create a watcher for Game instances
	watcher, err := dynamicClient.Resource(gamesGVR).Namespace("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error creating watcher: %v\n", err)
		os.Exit(1)
	}

  manager, err := manager.New()
  if err != nil {
    panic(err)
  }

	// Watch for events
	for event := range watcher.ResultChan() {
		obj, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		// Get details from the game instance
		name := obj.GetName()
		namespace := obj.GetNamespace()
		key := fmt.Sprintf("%s/%s", namespace, name)
    creationTimestamp := obj.GetCreationTimestamp()
		
    // Extract some details from the spec to display
    spec, found, err := unstructured.NestedMap(obj.Object, "spec")
    if err != nil || !found {
      fmt.Println("  Unable to extract spec details")
      continue
    }

		// Only notify for new games
		switch event.Type {
    case "ADDED":
      if time.Since(creationTimestamp.Time).Seconds() > 60 {
        continue
      }

      fmt.Printf("CRD Event detected: %s\n", key)
      fmt.Printf("Event Type: %s\n", event.Type)

      // Display some fields from the Game
      var gameType string
      var ok bool
      fmt.Printf("  Name: %s\n", name)
      
      if gameType, ok = spec["gameType"].(string); ok {
        fmt.Printf("  Type: %s\n", gameType)
      }
      
      chartPath := "./charts/" + gameType
      err = manager.Create(chartPath, name, namespace)
      if err != nil {
        fmt.Printf("Error creating resources: %v\n", err)
        continue
      }
    case "MODIFIED":
      fmt.Printf("CRD Event detected: %s\n", key)
      fmt.Printf("Event Type: %s\n", event.Type)

      fmt.Printf("  Name: %s\n", name)
      
      if gameType, ok := spec["gameType"].(string); ok {
        fmt.Printf("  Type: %s\n", gameType)
      }
      
      if players, ok := spec["players"].(int64); ok {
        fmt.Printf("  Players: %d\n", players)
      }
      err = manager.Update()
      if err != nil {
        fmt.Printf("Error updating resources: %v\n", err)
        continue
      }
    case "DELETED":
      fmt.Printf("CRD Event detected: %s\n", key)
      fmt.Printf("Event Type: %s\n", event.Type)

      var gameType string
      var ok bool

      // Display some fields from the Game
      fmt.Printf("  Name: %s\n", name)
      
      if gameType, ok = spec["gameType"].(string); ok {
        fmt.Printf("  Type: %s\n", gameType)
      }
      
      err = manager.Delete(name, namespace)
      if err != nil {
        fmt.Printf("Error deleting resources: %v\n", err)
        continue
      }
		}
	}
}
