package main

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
)

// ResourceQuery represents a query for a resource type
type ResourceQuery struct {
	Type   resource.Type
	IsList bool
}

// TreeNode represents a node in the resource tree
type TreeNode struct {
	ID       widget.TreeNodeID
	Type     string
	Label    string
	Resource map[string]interface{}
	Children []*TreeNode
}

// AppState holds the application state
type AppState struct {
	stateClient              state.State
	selectedResourceType     string
	treeRoot                 *TreeNode
	resourceTree             *widget.Tree
	resourceIDInput          *widget.Entry
	statusLabel              *widget.Label
	detailTitle              *widget.Label
	detailText               *widget.Entry
	currentResource          map[string]interface{}
	machineLinksContainer    *fyne.Container
	versionLinksContainer    *fyne.Container
	machineSetLinksContainer *fyne.Container
	machineRelatedLinksContainer *fyne.Container
	resourceActionsContainer *fyne.Container
}

// DetailComponents holds UI components for the detail pane
type DetailComponents struct {
	Title              *widget.Label
	Text               *widget.Entry
	MachineLinks       *fyne.Container
	VersionLinks       *fyne.Container
	MachineSetLinks    *fyne.Container
	MachineRelatedLinks *fyne.Container
	ResourceActions    *fyne.Container
}

// ResourceContext provides context for resource operations
type ResourceContext struct {
	AppState    *AppState
	StateClient state.State
	Ctx         context.Context
	Depth       int
}

// SettingsFormFields groups form field widgets for settings
type SettingsFormFields struct {
	EndpointEntry         *widget.Entry
	AuthSelect            *widget.Select
	ServiceAccountEntry   *widget.Entry
	OIDCIssuerEntry       *widget.Entry
	OIDCClientIDEntry     *widget.Entry
	OIDCClientSecretEntry *widget.Entry
	ShowEmptyLabelsCheck  *widget.Check
	IgnoreEnvCheck        *widget.Check
	GrpcDebugLevelSelect  *widget.Select
}
