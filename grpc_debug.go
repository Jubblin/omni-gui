package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
)

// grpcDebugLevel controls the level of gRPC debugging
// 0 = disabled, 1 = query dumps only, 2 = query + response dumps, 3 = query + response + UI display dumps
var grpcDebugLevel = 0

// setGrpcDebugLevel sets the gRPC debug level from command-line flag
func setGrpcDebugLevel(level int) {
	grpcDebugLevel = level
	if level > 0 {
		slog.Info("gRPC debugging enabled", "level", level)
	}
}

// debugQuery dumps query metadata as JSON when debug is enabled
func debugQuery(operation string, md resource.Metadata) {
	if grpcDebugLevel < 1 {
		return
	}

	slog.Info("gRPC Query",
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"operation", operation,
		"namespace", md.Namespace(),
		"type", string(md.Type()),
		"id", md.ID(),
		"version", md.Version().String())
}

// debugResponse dumps response data as JSON when debug level >= 2
func debugResponse(operation string, md resource.Metadata, res resource.Resource, err error) {
	if grpcDebugLevel < 2 {
		return
	}

	args := []interface{}{
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"operation", operation,
		"namespace", md.Namespace(),
		"type", string(md.Type()),
		"id", md.ID(),
	}

	if err != nil {
		args = append(args, "error", err.Error())
	} else if res != nil {
		args = append(args,
			"found", true,
			"resource_id", res.Metadata().ID(),
			"resource_version", res.Metadata().Version().String())
	} else {
		args = append(args, "found", false)
	}

	slog.Info("gRPC Response", args...)
}

// debugListResponse dumps list response data as JSON when debug level >= 2
func debugListResponse(operation string, md resource.Metadata, items []resource.Resource, err error) {
	if grpcDebugLevel < 2 {
		return
	}

	args := []interface{}{
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"operation", operation,
		"namespace", md.Namespace(),
		"type", string(md.Type()),
		"count", len(items),
	}

	if err != nil {
		args = append(args, "error", err.Error())
	} else {
		ids := make([]string, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.Metadata().ID())
		}
		args = append(args, "resource_ids", ids)
	}

	slog.Info("gRPC List Response", args...)
}

// debugUIDisplay dumps resource data when displayed in UI (debug level >= 3)
func debugUIDisplay(resourceData map[string]interface{}, resourceID, resourceType string) {
	if grpcDebugLevel < 3 {
		return
	}

	displayInfo := map[string]interface{}{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"context":      "ui_display",
		"resource_id":   resourceID,
		"resource_type": resourceType,
		"resource_data": resourceData,
	}

	// Write to a debug dump file
	debugDir := filepath.Join(os.TempDir(), "omni-api-debug")
	if err := os.MkdirAll(debugDir, 0755); err == nil {
		jsonData, err := json.MarshalIndent(displayInfo, "", "  ")
		if err == nil {
			filename := fmt.Sprintf("ui-display-%s-%s-%d.json", resourceType, resourceID, time.Now().Unix())
			filepath := filepath.Join(debugDir, filename)
			if err := os.WriteFile(filepath, jsonData, 0644); err == nil {
				slog.Info("gRPC UI Display Dump", "file", filepath, "resource_id", resourceID, "resource_type", resourceType)
			}
		}
	}

	slog.Info("gRPC UI Display",
		"timestamp", displayInfo["timestamp"],
		"context", displayInfo["context"],
		"resource_id", displayInfo["resource_id"],
		"resource_type", displayInfo["resource_type"],
		"resource_data", displayInfo["resource_data"])
}

// debugStateGet wraps state.Get with debugging
func debugStateGet(ctx context.Context, st state.State, md resource.Metadata) (resource.Resource, error) {
	debugQuery("Get", md)
	res, err := st.Get(ctx, md)
	debugResponse("Get", md, res, err)
	return res, err
}

// debugStateList wraps state.List with debugging
func debugStateList(ctx context.Context, st state.State, md resource.Metadata) (resource.List, error) {
	debugQuery("List", md)
	result, err := st.List(ctx, md)
	if err == nil {
		debugListResponse("List", md, result.Items, err)
	} else {
		debugListResponse("List", md, nil, err)
	}
	return result, err
}
