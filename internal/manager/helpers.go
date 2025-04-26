package manager

import (
	"encoding/json"
	"fmt"

	"github.com/Sackbuoy/gameserver-operator/internal/crds"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

// Convert map[string]any to GameServer struct.
func mapToGameServer(data map[string]any) (*crds.GameServer, error) {
	// First, marshal the map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Then unmarshal JSON to the struct
	var gameServer crds.GameServer
	if err := json.Unmarshal(jsonData, &gameServer); err != nil {
		return nil, err
	}

	return &gameServer, nil
}

func getInstalledCharts(actionConfig *action.Configuration) ([]*release.Release, error) {
	lister := action.NewList(actionConfig)
	lister.All = true
	lister.AllNamespaces = false
	lister.SetStateMask()

	releases, err := lister.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}

	return releases, nil
}
