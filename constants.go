package main

const (
	labelKeyMachineSet        = "omni.sidero.dev/machine-set"
	labelKeyCluster           = "omni.sidero.dev/cluster"
	resourceTypeFolder        = "resource-type-folder"
	clustersFolder            = "clusters-folder"
	machinesetsFolder         = "machinesets-folder"
	machinesFolder            = "machines-folder"
	clusterMachineLabelPrefix = "(ClusterMachine) %s"
)

// Version is the application version, set at build time
// This should be set via ldflags when building
var Version = "dev"
