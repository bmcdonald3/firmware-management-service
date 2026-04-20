package reconcilers

import (
"context"
"encoding/json"
"fmt"

v1 "github.com/bmcdonald3/fms/apis/example.fabrica.dev/v1"
"github.com/google/uuid"
"github.com/openchami/fabrica/pkg/fabrica"
)

// reconcileUpdateJob splits an UpdateJob into one UpdateTask per target node.
//
// Idempotency is ensured by listing existing UpdateTasks filtered by UpdateJobId
// before creating new ones, so re-reconciliation never duplicates tasks.
func (r *UpdateJobReconciler) reconcileUpdateJob(ctx context.Context, res *v1.UpdateJob) error {
if res.Status.Phase == "Complete" || res.Status.Phase == "Failed" {
return nil
}

// List all existing UpdateTasks to check which nodes already have a task.
items, err := r.Client.List(ctx, "UpdateTask")
if err != nil {
return fmt.Errorf("failed to list UpdateTasks: %w", err)
}

// Build a set of nodes that already have a task for this job.
covered := make(map[string]struct{})
for _, item := range items {
data, err := json.Marshal(item)
if err != nil {
continue
}
var task v1.UpdateTask
if err := json.Unmarshal(data, &task); err != nil {
continue
}
if task.Spec.UpdateJobId == res.GetUID() {
covered[task.Spec.TargetNode] = struct{}{}
}
}

// Create a task for each uncovered target node.
for _, node := range res.Spec.TargetNodes {
if _, exists := covered[node]; exists {
continue
}
task := &v1.UpdateTask{
APIVersion: "example.fabrica.dev/v1",
Kind:       "UpdateTask",
Metadata: fabrica.Metadata{
Name: fmt.Sprintf("%s-%s", res.Metadata.Name, node),
UID:  uuid.NewString(),
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
		r.Logger.Infof("Created UpdateTask for node %s (job %s)", node, res.GetUID())
		// Emit a created event so the controller queues the task for reconciliation.
		if err := r.EmitEvent(ctx, "fms.resource.updatetask.created", task); err != nil {
			r.Logger.Warnf("Failed to emit event for UpdateTask %s: %v", task.GetUID(), err)
		}
}

// Aggregate child task states to determine overall job phase.
items, err = r.Client.List(ctx, "UpdateTask")
if err != nil {
return fmt.Errorf("failed to re-list UpdateTasks: %w", err)
}

total, success, failed := 0, 0, 0
for _, item := range items {
data, err := json.Marshal(item)
if err != nil {
continue
}
var task v1.UpdateTask
if err := json.Unmarshal(data, &task); err != nil {
continue
}
if task.Spec.UpdateJobId != res.GetUID() {
continue
}
total++
switch task.Status.State {
case "Success":
success++
case "Failed":
failed++
}
}

switch {
case total == 0:
res.Status.Phase = "Pending"
case failed > 0:
res.Status.Phase = "Failed"
res.Status.Message = fmt.Sprintf("%d/%d tasks failed", failed, total)
case success == total:
res.Status.Phase = "Complete"
res.Status.Message = fmt.Sprintf("All %d tasks completed successfully", total)
default:
res.Status.Phase = "Running"
res.Status.Message = fmt.Sprintf("%d/%d tasks complete", success, total)
}

if err := r.Client.Update(ctx, res); err != nil {
return fmt.Errorf("failed to update UpdateJob status: %w", err)
}

r.Logger.Infof("UpdateJob %s phase: %s", res.GetUID(), res.Status.Phase)
return nil
}