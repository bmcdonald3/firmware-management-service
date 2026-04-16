// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
// This file contains user-customizable reconciliation logic for UpdateJob.
//
// ⚠️ This file is safe to edit - it will NOT be overwritten by code generation.
package reconcilers

import (
"context"
"fmt"

v1 "github.com/bmcdonald3/fms/apis/firmware.management.io/v1"
"github.com/openchami/fabrica/pkg/fabrica"
)

// reconcileUpdateJob is the splitter: for each targetNode in the UpdateJob spec
// it ensures a child UpdateTask exists, then aggregates child states back to
// the parent UpdateJob status.
func (r *UpdateJobReconciler) reconcileUpdateJob(ctx context.Context, res *v1.UpdateJob) error {
r.Logger.Infof("Reconciling UpdateJob %s (nodes=%v)", res.GetUID(), res.Spec.TargetNodes)

// --- 1. List all UpdateTasks and filter in-memory (Fabrica has no server-side field filter) ---
rawItems, err := r.Client.List(ctx, "UpdateTask")
if err != nil {
return fmt.Errorf("failed to list UpdateTasks: %w", err)
}

// Build a set of (updateJobId, targetNode) tuples that already exist.
type taskKey struct{ jobID, node string }
existing := make(map[taskKey]bool)
var childTasks []*v1.UpdateTask

for _, item := range rawItems {
task, ok := item.(*v1.UpdateTask)
if !ok {
continue
}
if task.Spec.UpdateJobId == res.GetUID() {
existing[taskKey{task.Spec.UpdateJobId, task.Spec.TargetNode}] = true
childTasks = append(childTasks, task)
}
}

// --- 2. Create missing UpdateTask resources for each target node ---
for _, node := range res.Spec.TargetNodes {
key := taskKey{res.GetUID(), node}
if existing[key] {
continue
}

task := &v1.UpdateTask{
APIVersion: "firmware.management.io/v1",
Kind:       "UpdateTask",
Metadata: fabrica.Metadata{
Name: fmt.Sprintf("%s-%s", res.Metadata.Name, node),
},
Spec: v1.UpdateTaskSpec{
TargetNode:      node,
TargetComponent: res.Spec.TargetComponent,
FirmwareRef:     res.Spec.FirmwareRef,
UpdateJobId:     res.GetUID(),
},
Status: v1.UpdateTaskStatus{
State: "Pending",
},
}
if err := r.Client.Create(ctx, task); err != nil {
return fmt.Errorf("failed to create UpdateTask for node %s: %w", node, err)
}
r.Logger.Infof("Created UpdateTask %s for node %s", task.Metadata.Name, node)
childTasks = append(childTasks, task)
}

// --- 3. Aggregate child states to determine parent UpdateJob phase ---
total := len(childTasks)
if total == 0 {
res.Status.Phase = "Pending"
res.Status.Message = "No target nodes"
return r.Client.Update(ctx, res)
}

counts := map[string]int{"Pending": 0, "Running": 0, "Success": 0, "Failed": 0}
for _, t := range childTasks {
state := t.Status.State
if state == "" {
state = "Pending"
}
counts[state]++
}

switch {
case counts["Failed"] > 0:
res.Status.Phase = "Failed"
res.Status.Message = fmt.Sprintf("%d/%d tasks failed", counts["Failed"], total)
case counts["Running"] > 0:
res.Status.Phase = "Running"
res.Status.Message = fmt.Sprintf("%d/%d tasks running", counts["Running"], total)
case counts["Success"] == total:
res.Status.Phase = "Complete"
res.Status.Message = "All tasks succeeded"
default:
res.Status.Phase = "Pending"
res.Status.Message = fmt.Sprintf("%d/%d tasks pending", counts["Pending"], total)
}

if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("failed to update UpdateJob status: %w", err)
}
r.Logger.Infof("UpdateJob %s phase=%s", res.GetUID(), res.Status.Phase)
return nil
}