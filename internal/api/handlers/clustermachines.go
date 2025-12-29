package handlers

import (
	"log"
	"net/http"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/gin-gonic/gin"
	omniresources "github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
)

// ClusterMachineResponse represents the cluster machine information returned by the API
type ClusterMachineResponse struct {
	ID                string            `json:"id"`
	Namespace         string            `json:"namespace"`
	MachineID         string            `json:"machine_id,omitempty"`
	KubernetesVersion string            `json:"kubernetes_version,omitempty"`
	ManagementAddress string            `json:"management_address,omitempty"`
	Connected         bool              `json:"connected,omitempty"`
	UseGrpcTunnel     bool              `json:"use_grpc_tunnel,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	// Status fields (from MachineStatus resource) - same as MachineResponse
	Hostname          string            `json:"hostname,omitempty"`
	Platform          string            `json:"platform,omitempty"`
	Arch              string            `json:"arch,omitempty"`
	TalosVersion      string            `json:"talos_version,omitempty"`
	Role              string            `json:"role,omitempty"`
	Maintenance       bool              `json:"maintenance,omitempty"`
	LastError         string            `json:"last_error,omitempty"`
	Links             map[string]string `json:"_links,omitempty"`
}

// ClusterMachineHandler handles cluster machine requests
type ClusterMachineHandler struct {
	state state.State
}

// NewClusterMachineHandler creates a new ClusterMachineHandler
func NewClusterMachineHandler(s state.State) *ClusterMachineHandler {
	return &ClusterMachineHandler{state: s}
}

// ListClusterMachines godoc
// @Summary      List all cluster machines
// @Description  Get a list of all cluster machines in Omni
// @Tags         clustermachines
// @Produce      json
// @Param        cluster   query     string  false  "Filter by cluster ID"
// @Success      200  {array}   ClusterMachineResponse
// @Failure      500  {object}  map[string]string
// @Router       /clustermachines [get]
func (h *ClusterMachineHandler) ListClusterMachines(c *gin.Context) {
	st := h.state

	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, "", resource.VersionUndefined)

	items, err := st.List(c.Request.Context(), md)
	if err != nil {
		log.Printf("Error listing cluster machines: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clusterFilter := c.Query("cluster")

	var clusterMachines []ClusterMachineResponse
	for _, item := range items.Items {
		cm, ok := item.(*omni.ClusterMachine)
		if !ok {
			log.Printf("Warning: resource is not a cluster machine: %T", item)
			continue
		}

		// Filter by cluster if specified
		if clusterFilter != "" {
			if clusterID, ok := cm.Metadata().Labels().Get("omni.sidero.dev/cluster"); !ok || clusterID != clusterFilter {
				continue
			}
		}

		spec := cm.TypedSpec().Value
		clusterMachineID := cm.Metadata().ID()
		resp := ClusterMachineResponse{
			ID:        clusterMachineID,
			Namespace: cm.Metadata().Namespace(),
			Labels:    make(map[string]string),
			Links: map[string]string{
				"self": buildURL(c, "/api/v1/clustermachines/"+clusterMachineID),
			},
		}

		// The ClusterMachine ID is typically the machine ID
		resp.MachineID = clusterMachineID
		resp.Links["machine"] = buildURL(c, "/api/v1/machines/"+clusterMachineID)
		
		if spec.KubernetesVersion != "" {
			resp.KubernetesVersion = spec.KubernetesVersion
		}

		// Collect all metadata labels
		for key, value := range cm.Metadata().Labels().Raw() {
			resp.Labels[key] = value
		}

		// Try to fetch Machine resource to get management address, connected status, etc.
		machineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, clusterMachineID, resource.VersionUndefined)
		if machineRes, err := st.Get(c.Request.Context(), machineMD); err == nil {
			if m, ok := machineRes.(*omni.Machine); ok {
				resp.ManagementAddress = m.TypedSpec().Value.ManagementAddress
				resp.Connected = m.TypedSpec().Value.Connected
				resp.UseGrpcTunnel = m.TypedSpec().Value.UseGrpcTunnel
			}
		}

		// Try to fetch and include machine status information (same as MachineResponse)
		statusMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, clusterMachineID, resource.VersionUndefined)
		if statusRes, err := st.Get(c.Request.Context(), statusMD); err == nil {
			if ms, ok := statusRes.(*omni.MachineStatus); ok {
				statusSpec := ms.TypedSpec().Value
				resp.TalosVersion = statusSpec.TalosVersion
				resp.Role = statusSpec.Role.String()
				resp.Maintenance = statusSpec.Maintenance
				if statusSpec.LastError != "" {
					resp.LastError = statusSpec.LastError
				}
				if statusSpec.Network != nil {
					resp.Hostname = statusSpec.Network.Hostname
				}
				if statusSpec.PlatformMetadata != nil {
					resp.Platform = statusSpec.PlatformMetadata.Platform
				}
				if statusSpec.Hardware != nil {
					resp.Arch = statusSpec.Hardware.Arch
				}
			}
		}

		// Try to find cluster ID from labels
		if clusterID, ok := cm.Metadata().Labels().Get("omni.sidero.dev/cluster"); ok {
			resp.Links["cluster"] = buildURL(c, "/api/v1/clusters/"+clusterID)
		}

		// Add links to related resources (same as MachineResponse)
		resp.Links["labels"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/labels")
		resp.Links["extensions"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/extensions")
		resp.Links["upgrade-status"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/upgrade-status")
		resp.Links["metrics"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/metrics")
		resp.Links["status"] = buildURL(c, "/api/v1/clustermachines/"+clusterMachineID+"/status")
		resp.Links["config-status"] = buildURL(c, "/api/v1/clustermachines/"+clusterMachineID+"/config-status")
		resp.Links["talos-version"] = buildURL(c, "/api/v1/clustermachines/"+clusterMachineID+"/talos-version")
		resp.Links["machine-status"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/status")

		clusterMachines = append(clusterMachines, resp)
	}

	c.JSON(http.StatusOK, clusterMachines)
}

// GetClusterMachine godoc
// @Summary      Get a single cluster machine
// @Description  Get detailed information about a specific cluster machine
// @Tags         clustermachines
// @Produce      json
// @Param        id   path      string  true  "Cluster Machine ID"
// @Success      200  {object}  ClusterMachineResponse
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /clustermachines/{id} [get]
func (h *ClusterMachineHandler) GetClusterMachine(c *gin.Context) {
	id := c.Param("id")
	st := h.state

	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, id, resource.VersionUndefined)

	res, err := st.Get(c.Request.Context(), md)
	if err != nil {
		log.Printf("Error getting cluster machine %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cm, ok := res.(*omni.ClusterMachine)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error: unexpected resource type"})
		return
	}

	spec := cm.TypedSpec().Value
	clusterMachineID := cm.Metadata().ID()
	resp := ClusterMachineResponse{
		ID:        clusterMachineID,
		Namespace: cm.Metadata().Namespace(),
		Labels:    make(map[string]string),
		Links: map[string]string{
			"self": buildURL(c, "/api/v1/clustermachines/"+clusterMachineID),
		},
	}

	// The ClusterMachine ID is typically the machine ID
	resp.MachineID = clusterMachineID
	resp.Links["machine"] = buildURL(c, "/api/v1/machines/"+clusterMachineID)
	
	if spec.KubernetesVersion != "" {
		resp.KubernetesVersion = spec.KubernetesVersion
	}

	// Collect all metadata labels
	for key, value := range cm.Metadata().Labels().Raw() {
		resp.Labels[key] = value
	}

	// Try to fetch Machine resource to get management address, connected status, etc.
	machineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, clusterMachineID, resource.VersionUndefined)
	if machineRes, err := st.Get(c.Request.Context(), machineMD); err == nil {
		if m, ok := machineRes.(*omni.Machine); ok {
			resp.ManagementAddress = m.TypedSpec().Value.ManagementAddress
			resp.Connected = m.TypedSpec().Value.Connected
			resp.UseGrpcTunnel = m.TypedSpec().Value.UseGrpcTunnel
		}
	}

	// Try to fetch and include machine status information (same as MachineResponse)
	statusMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, clusterMachineID, resource.VersionUndefined)
	if statusRes, err := st.Get(c.Request.Context(), statusMD); err == nil {
		if ms, ok := statusRes.(*omni.MachineStatus); ok {
			statusSpec := ms.TypedSpec().Value
			resp.TalosVersion = statusSpec.TalosVersion
			resp.Role = statusSpec.Role.String()
			resp.Maintenance = statusSpec.Maintenance
			if statusSpec.LastError != "" {
				resp.LastError = statusSpec.LastError
			}
			if statusSpec.Network != nil {
				resp.Hostname = statusSpec.Network.Hostname
			}
			if statusSpec.PlatformMetadata != nil {
				resp.Platform = statusSpec.PlatformMetadata.Platform
			}
			if statusSpec.Hardware != nil {
				resp.Arch = statusSpec.Hardware.Arch
			}
		}
	}

	// Try to find cluster ID from labels
	if clusterID, ok := cm.Metadata().Labels().Get("omni.sidero.dev/cluster"); ok {
		resp.Links["cluster"] = buildURL(c, "/api/v1/clusters/"+clusterID)
	}

	// Add links to related resources (same as MachineResponse)
	resp.Links["labels"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/labels")
	resp.Links["extensions"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/extensions")
	resp.Links["upgrade-status"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/upgrade-status")
	resp.Links["metrics"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/metrics")
	resp.Links["status"] = buildURL(c, "/api/v1/clustermachines/"+clusterMachineID+"/status")
	resp.Links["config-status"] = buildURL(c, "/api/v1/clustermachines/"+clusterMachineID+"/config-status")
	resp.Links["talos-version"] = buildURL(c, "/api/v1/clustermachines/"+clusterMachineID+"/talos-version")
	resp.Links["machine-status"] = buildURL(c, "/api/v1/machines/"+clusterMachineID+"/status")

	c.JSON(http.StatusOK, resp)
}
