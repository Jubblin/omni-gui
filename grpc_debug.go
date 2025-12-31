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

	queryInfo := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"operation":  operation,
		"namespace":  md.Namespace(),
		"type":       string(md.Type()),
		"id":         md.ID(),
		"version":    md.Version().String(),
	}

	jsonData, err := json.MarshalIndent(queryInfo, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal query debug info", "error", err)
		return
	}

	slog.Info("gRPC Query", "query", string(jsonData))
}

// debugResponse dumps response data as JSON when debug level >= 2
func debugResponse(operation string, md resource.Metadata, res resource.Resource, err error) {
	if grpcDebugLevel < 2 {
		return
	}

	responseInfo := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"operation": operation,
		"namespace": md.Namespace(),
		"type":      string(md.Type()),
		"id":        md.ID(),
		"error":     nil,
	}

	if err != nil {
		responseInfo["error"] = err.Error()
	} else if res != nil {
		responseInfo["found"] = true
		responseInfo["resource_id"] = res.Metadata().ID()
		responseInfo["resource_version"] = res.Metadata().Version().String()
	} else {
		responseInfo["found"] = false
	}

	jsonData, err := json.MarshalIndent(responseInfo, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal response debug info", "error", err)
		return
	}

	slog.Info("gRPC Response", "response", string(jsonData))
}

// debugListResponse dumps list response data as JSON when debug level >= 2
func debugListResponse(operation string, md resource.Metadata, items []resource.Resource, err error) {
	if grpcDebugLevel < 2 {
		return
	}

	responseInfo := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"operation": operation,
		"namespace": md.Namespace(),
		"type":      string(md.Type()),
		"error":     nil,
		"count":     len(items),
	}

	if err != nil {
		responseInfo["error"] = err.Error()
	} else {
		ids := make([]string, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.Metadata().ID())
		}
		responseInfo["resource_ids"] = ids
	}

	jsonData, err := json.MarshalIndent(responseInfo, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal list response debug info", "error", err)
		return
	}

	slog.Info("gRPC List Response", "response", string(jsonData))
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

	jsonData, err := json.MarshalIndent(displayInfo, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal UI display debug info", "error", err)
		return
	}

	// Write to a debug dump file
	debugDir := filepath.Join(os.TempDir(), "omni-api-debug")
	if err := os.MkdirAll(debugDir, 0755); err == nil {
		filename := fmt.Sprintf("ui-display-%s-%s-%d.json", resourceType, resourceID, time.Now().Unix())
		filepath := filepath.Join(debugDir, filename)
		if err := os.WriteFile(filepath, jsonData, 0644); err == nil {
			slog.Info("gRPC UI Display Dump", "file", filepath, "resource_id", resourceID, "resource_type", resourceType)
		}
	}

	slog.Info("gRPC UI Display", "display", string(jsonData))
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
