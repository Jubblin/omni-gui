package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/resource/protobuf"
	"github.com/cosi-project/runtime/pkg/resource/typed"
	"github.com/gin-gonic/gin"
	"github.com/siderolabs/omni/client/api/omni/specs"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMachineStatusHandler_GetMachineStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockState := new(MockState)

	m := typed.NewResource[omni.MachineSpec, omni.MachineExtension](
		resource.NewMetadata("default", omni.MachineType, "machine-1", resource.VersionUndefined),
		protobuf.NewResourceSpec(&specs.MachineSpec{
			ManagementAddress: "192.168.1.10",
			Connected:         true,
		}),
	)

	ms := typed.NewResource[omni.MachineStatusSpec, omni.MachineStatusExtension](
		resource.NewMetadata("default", omni.MachineStatusType, "machine-1", resource.VersionUndefined),
		protobuf.NewResourceSpec(&specs.MachineStatusSpec{
			TalosVersion: "v1.5.0",
			Role:         specs.MachineStatusSpec_CONTROL_PLANE,
			Maintenance:  false,
			Network: &specs.MachineStatusSpec_NetworkStatus{
				Hostname: "talos-node-1",
			},
			PlatformMetadata: &specs.MachineStatusSpec_PlatformMetadata{
				Platform: "metal",
			},
			Hardware: &specs.MachineStatusSpec_HardwareStatus{
				Arch: "amd64",
			},
		}),
	)

	mockState.On("Get", mock.Anything, mock.MatchedBy(func(md resource.Pointer) bool {
		return md.Type() == omni.MachineType
	}), mock.Anything).Return(m, nil)

	mockState.On("Get", mock.Anything, mock.MatchedBy(func(md resource.Pointer) bool {
		return md.Type() == omni.MachineStatusType
	}), mock.Anything).Return(ms, nil)

	handler := NewMachineStatusHandler(mockState)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "machine-1"}}
	c.Request, _ = http.NewRequest("GET", "/machines/machine-1/status", nil)

	handler.GetMachineStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp MachineResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "machine-1", resp.ID)
	assert.NotNil(t, resp.Status, "Status should be populated")
	if resp.Status != nil {
		assert.Equal(t, "v1.5.0", resp.Status.TalosVersion)
		assert.Equal(t, "CONTROL_PLANE", resp.Status.Role)
		assert.Equal(t, "talos-node-1", resp.Status.Hostname)
		assert.Equal(t, "metal", resp.Status.Platform)
		assert.Equal(t, "amd64", resp.Status.Arch)
	}
}
