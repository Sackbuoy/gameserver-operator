package crds

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GameServer defines the schema for the GameServer custom resource.
type GameServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GameServerSpec   `json:"spec,omitempty"`
	Status GameServerStatus `json:"status,omitempty"`
}

// GameServerSpec defines the desired state of a GameServer.
type GameServerSpec struct {
	// Type of game server (minecraft, valheim, etc.)
	GameType string `json:"gameType"`

	// HelmChart contains the details of the Helm chart to deploy
	HelmChart HelmChart `json:"helmChart,omitempty"`

	// Resources describes the compute resources allocated to the game server
	Resources ResourceRequirements `json:"resources,omitempty"`

	// Persistence configuration for the game server
	Persistence *PersistenceConfig `json:"persistence,omitempty"`

	// Networking configuration for the game server
	Networking *NetworkingConfig `json:"networking,omitempty"`
}

// HelmChart contains details about a Helm chart to deploy.
type HelmChart struct {
	// Repository is the URL of the Helm chart repository
	Repository string `json:"repository"`

	// Name of the Helm chart
	Name string `json:"name"`

	// Version of the Helm chart to use
	Version string `json:"version"`

	// ValuesOverride contains Helm chart values to override, stored as a YAML string
	ValuesOverride string `json:"valuesOverride,omitempty"`

	// Timeout for Helm operations in seconds
	Timeout int `json:"timeout,omitempty"`
}

// ResourceRequirements describes the compute resource requirements.
type ResourceRequirements struct {
	// Requests describes the minimum resource requirements
	Requests *ResourceList `json:"requests,omitempty"`

	// Limits describes the maximum resource requirements
	Limits *ResourceList `json:"limits,omitempty"`
}

// ResourceList contains resource quantities.
type ResourceList struct {
	// CPU resource request/limit (e.g., '500m', '1')
	CPU string `json:"cpu,omitempty"`

	// Memory resource request/limit (e.g., '1Gi')
	Memory string `json:"memory,omitempty"`

	// EphemeralStorage request/limit (e.g., '10Gi')
	EphemeralStorage string `json:"ephemeralStorage,omitempty"`
}

// PersistenceConfig defines persistent storage configuration.
type PersistenceConfig struct {
	// Whether to enable persistent storage
	Enabled bool `json:"enabled,omitempty"`

	// Size of persistent volume (e.g., '10Gi')
	Size string `json:"size,omitempty"`

	// StorageClass for the PVC
	StorageClass string `json:"storageClass,omitempty"`
}

// NetworkingConfig defines networking configuration.
type NetworkingConfig struct {
	// Type of service (ClusterIP, NodePort, LoadBalancer)
	Type string `json:"type,omitempty"`

	// Ports to expose
	Ports []PortConfig `json:"ports,omitempty"`

	// Annotations for the service
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PortConfig defines a port configuration.
type PortConfig struct {
	// Name of the port
	Name string `json:"name,omitempty"`

	// Port number
	Port int32 `json:"port"`

	// Target port number (defaults to port)
	TargetPort int32 `json:"targetPort,omitempty"`

	// Protocol for this port (TCP, UDP)
	Protocol string `json:"protocol,omitempty"`

	// Node port when type is NodePort
	NodePort int32 `json:"nodePort,omitempty"`
}

// GameServerStatus defines the observed state of a GameServer.
type GameServerStatus struct {
	// Current phase of the game server (Pending, Deploying, Running, Failed, etc.)
	Phase string `json:"phase,omitempty"`

	// Human-readable message about the current state
	Message string `json:"message,omitempty"`

	// HelmRelease contains information about the Helm release
	HelmRelease *HelmReleaseStatus `json:"helmRelease,omitempty"`

	// Deployment contains information about the deployment
	Deployment *DeploymentStatus `json:"deployment,omitempty"`

	// Networking contains information about the service
	Networking *NetworkingStatus `json:"networking,omitempty"`

	// Conditions is a list of current conditions
	Conditions []GameServerCondition `json:"conditions,omitempty"`

	// LastUpdated is the last time the status was updated
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

// HelmReleaseStatus contains information about a Helm release.
type HelmReleaseStatus struct {
	// Name of the Helm release
	Name string `json:"name,omitempty"`

	// Version of the Helm release
	Version int `json:"version,omitempty"`

	// LastDeployed is the last time the Helm release was deployed
	LastDeployed *metav1.Time `json:"lastDeployed,omitempty"`
}

// DeploymentStatus contains information about a deployment.
type DeploymentStatus struct {
	// Whether the deployment is available
	Available bool `json:"available,omitempty"`

	// Current number of replicas
	Replicas int32 `json:"replicas,omitempty"`

	// Number of ready replicas
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Number of updated replicas
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`
}

// NetworkingStatus contains information about the service.
type NetworkingStatus struct {
	// Type of service created
	ServiceType string `json:"serviceType,omitempty"`

	// Cluster IP of the service
	ClusterIP string `json:"clusterIP,omitempty"`

	// External IP for LoadBalancer service
	ExternalIP string `json:"externalIP,omitempty"`

	// Ports exposed by the service
	Ports []PortStatus `json:"ports,omitempty"`
}

// PortStatus contains information about an exposed port.
type PortStatus struct {
	// Name of the port
	Name string `json:"name,omitempty"`

	// Port number
	Port int32 `json:"port,omitempty"`

	// Target port number
	TargetPort int32 `json:"targetPort,omitempty"`

	// Node port
	NodePort int32 `json:"nodePort,omitempty"`

	// Protocol for this port
	Protocol string `json:"protocol,omitempty"`
}

// GameServerCondition contains condition information for a GameServer.
type GameServerCondition struct {
	// Type of condition
	Type string `json:"type"`

	// Status of the condition (True, False, Unknown)
	Status string `json:"status"`

	// LastTransitionTime is the last time the condition transitioned
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason for the condition's last transition
	Reason string `json:"reason,omitempty"`

	// Message about the last transition
	Message string `json:"message,omitempty"`
}

// GameServerList contains a list of GameServer resources.
type GameServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GameServer `json:"items"`
}
