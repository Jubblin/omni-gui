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

// MachineStatusInfo represents machine status information from MachineStatus resource
type MachineStatusInfo struct {
	Hostname     string `json:"hostname,omitempty"`
	Platform     string `json:"platform,omitempty"`
	Arch         string `json:"arch,omitempty"`
	TalosVersion string `json:"talos_version,omitempty"`
	Role         string `json:"role,omitempty"`
	Maintenance  bool   `json:"maintenance,omitempty"`
	LastError    string `json:"last_error,omitempty"`
}

// MachineResponse represents the machine information returned by the API
type MachineResponse struct {
	ID                string            `json:"id"`
	Namespace         string            `json:"namespace"`
	ManagementAddress string            `json:"management_address"`
	Connected         bool              `json:"connected"`
	UseGrpcTunnel     bool              `json:"use_grpc_tunnel"`
	Labels            map[string]string `json:"labels,omitempty"`
	Status            *MachineStatusInfo `json:"status,omitempty"`
	Links             map[string]string  `json:"_links,omitempty"`
}

// MachineHandler handles machine requests
type MachineHandler struct {
	state state.State
}

// NewMachineHandler creates a new MachineHandler
func NewMachineHandler(s state.State) *MachineHandler {
	return &MachineHandler{state: s}
}

// ListMachines godoc
// @Summary      List all machines
// @Description  Get a list of all machines in Omni
// @Tags         machines
// @Produce      json
// @Success      200  {array}   MachineResponse
// @Failure      500  {object}  map[string]string
// @Router       /machines [get]
func (h *MachineHandler) ListMachines(c *gin.Context) {
	st := h.state

	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, "", resource.VersionUndefined)
	
	items, err := st.List(c.Request.Context(), md)
	if err != nil {
		log.Printf("Error listing machines: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var machines []MachineResponse
	for _, item := range items.Items {
		// Try to cast to the typed machine resource
		m, ok := item.(*omni.Machine)
		if !ok {
			log.Printf("Warning: resource is not a machine: %T", item)
			continue
		}

		machineID := m.Metadata().ID()
		resp := MachineResponse{
			ID:                machineID,
			Namespace:         m.Metadata().Namespace(),
			ManagementAddress: m.TypedSpec().Value.ManagementAddress,
			Connected:         m.TypedSpec().Value.Connected,
			UseGrpcTunnel:     m.TypedSpec().Value.UseGrpcTunnel,
			Labels:            make(map[string]string),
			Links: map[string]string{
				"self": buildURL(c, "/api/v1/machines/"+machineID),
			},
		}

		// Collect all metadata labels
		for key, value := range m.Metadata().Labels().Raw() {
			resp.Labels[key] = value
		}

		// Try to fetch and include machine status information
		statusMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, machineID, resource.VersionUndefined)
		if statusRes, err := st.Get(c.Request.Context(), statusMD); err == nil {
			if ms, ok := statusRes.(*omni.MachineStatus); ok {
				spec := ms.TypedSpec().Value
				statusInfo := &MachineStatusInfo{
					TalosVersion: spec.TalosVersion,
					Role:         spec.Role.String(),
					Maintenance:  spec.Maintenance,
				}
				if spec.LastError != "" {
					statusInfo.LastError = spec.LastError
				}
				if spec.Network != nil {
					statusInfo.Hostname = spec.Network.Hostname
				}
				if spec.PlatformMetadata != nil {
					statusInfo.Platform = spec.PlatformMetadata.Platform
				}
				if spec.Hardware != nil {
					statusInfo.Arch = spec.Hardware.Arch
				}
				resp.Status = statusInfo
			}
		}

		// Add links based on labels
		if clusterID, ok := m.Metadata().Labels().Get("omni.sidero.dev/cluster"); ok {
			resp.Links["cluster"] = buildURL(c, "/api/v1/clusters/"+clusterID)
		}
		
		// Add links to related resources (will be populated if handlers exist)
		resp.Links["labels"] = buildURL(c, "/api/v1/machines/"+machineID+"/labels")
		resp.Links["extensions"] = buildURL(c, "/api/v1/machines/"+machineID+"/extensions")
		resp.Links["upgrade-status"] = buildURL(c, "/api/v1/machines/"+machineID+"/upgrade-status")
		resp.Links["metrics"] = buildURL(c, "/api/v1/machines/"+machineID+"/metrics")

		machines = append(machines, resp)
	}

	c.JSON(http.StatusOK, machines)
}

// GetMachine godoc
// @Summary      Get a single machine
// @Description  Get detailed information about a specific machine
// @Tags         machines
// @Produce      json
// @Param        id   path      string  true  "Machine ID"
// @Success      200  {object}  MachineResponse
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /machines/{id} [get]
func (h *MachineHandler) GetMachine(c *gin.Context) {
	id := c.Param("id")
	st := h.state

	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, id, resource.VersionUndefined)

	res, err := st.Get(c.Request.Context(), md)
	if err != nil {
		log.Printf("Error getting machine %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	m, ok := res.(*omni.Machine)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error: unexpected resource type"})
		return
	}

	machineID := m.Metadata().ID()
	resp := MachineResponse{
		ID:                machineID,
		Namespace:         m.Metadata().Namespace(),
		ManagementAddress: m.TypedSpec().Value.ManagementAddress,
		Connected:         m.TypedSpec().Value.Connected,
		UseGrpcTunnel:     m.TypedSpec().Value.UseGrpcTunnel,
		Labels:            make(map[string]string),
		Links: map[string]string{
			"self": buildURL(c, "/api/v1/machines/"+machineID),
		},
	}

	// Collect all metadata labels
	for key, value := range m.Metadata().Labels().Raw() {
		resp.Labels[key] = value
	}

	// Try to fetch and include machine status information
	statusMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, machineID, resource.VersionUndefined)
	if statusRes, err := st.Get(c.Request.Context(), statusMD); err == nil {
		if ms, ok := statusRes.(*omni.MachineStatus); ok {
			spec := ms.TypedSpec().Value
			statusInfo := &MachineStatusInfo{
				TalosVersion: spec.TalosVersion,
				Role:         spec.Role.String(),
				Maintenance:  spec.Maintenance,
			}
			if spec.LastError != "" {
				statusInfo.LastError = spec.LastError
			}
			if spec.Network != nil {
				statusInfo.Hostname = spec.Network.Hostname
			}
			if spec.PlatformMetadata != nil {
				statusInfo.Platform = spec.PlatformMetadata.Platform
			}
			if spec.Hardware != nil {
				statusInfo.Arch = spec.Hardware.Arch
			}
			resp.Status = statusInfo
		}
	}

	// Add links based on labels
	if clusterID, ok := m.Metadata().Labels().Get("omni.sidero.dev/cluster"); ok {
		resp.Links["cluster"] = buildURL(c, "/api/v1/clusters/"+clusterID)
	}
	
	// Add links to related resources
	resp.Links["labels"] = buildURL(c, "/api/v1/machines/"+machineID+"/labels")
	resp.Links["extensions"] = buildURL(c, "/api/v1/machines/"+machineID+"/extensions")
	resp.Links["upgrade-status"] = buildURL(c, "/api/v1/machines/"+machineID+"/upgrade-status")
	resp.Links["metrics"] = buildURL(c, "/api/v1/machines/"+machineID+"/metrics")

	c.JSON(http.StatusOK, resp)
}

