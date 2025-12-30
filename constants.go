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

// Version is set at build time via ldflags
var Version = "dev"
