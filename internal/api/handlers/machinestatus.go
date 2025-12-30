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

// MachineStatusResponse represents the machine status information returned by the API
type MachineStatusResponse struct {
	ID          string            `json:"id"`
	Namespace   string            `json:"namespace"`
	Hostname    string            `json:"hostname,omitempty"`
	Platform    string            `json:"platform,omitempty"`
	Arch        string            `json:"arch,omitempty"`
	TalosVersion string           `json:"talos_version,omitempty"`
	Links       map[string]string `json:"_links,omitempty"`
}

// MachineStatusHandler handles machine status requests
type MachineStatusHandler struct {
	state state.State
}

// NewMachineStatusHandler creates a new MachineStatusHandler
func NewMachineStatusHandler(s state.State) *MachineStatusHandler {
	return &MachineStatusHandler{state: s}
}

// GetMachineStatus godoc
// @Summary      Get machine status (deprecated: use GET /machines/{id} instead)
// @Description  Get detailed status information about a specific machine. This endpoint is deprecated - status information is now included in the main machine endpoint.
// @Tags         machines
// @Produce      json
// @Param        id   path      string  true  "Machine ID"
// @Success      200  {object}  MachineResponse
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /machines/{id}/status [get]
// @Deprecated
func (h *MachineStatusHandler) GetMachineStatus(c *gin.Context) {
	id := c.Param("id")
	st := h.state

	// Fetch the machine resource
	machineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, id, resource.VersionUndefined)
	machineRes, err := st.Get(c.Request.Context(), machineMD)
	if err != nil {
		log.Printf("Error getting machine %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "machine not found"})
		return
	}

	m, ok := machineRes.(*omni.Machine)
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

	// Fetch and include machine status information
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
