package storage

import (
	"context"
	"fmt"

	"github.com/bmcdonald3/firmware-management-service/fms/internal/storage/ent"
	"github.com/bmcdonald3/firmware-management-service/fms/internal/storage/ent/label"
	entresource "github.com/bmcdonald3/firmware-management-service/fms/internal/storage/ent/resource"

	v1 "github.com/bmcdonald3/firmware-management-service/fms/apis/firmware.management.io/v1"
)

// ensureEntClient verifies the ent client has been initialized
func ensureEntClient() {
	if entClient == nil {
		panic("ent client not initialized: call SetEntClient in main.go before using storage")
	}
}

// QueryResources returns a generic query builder for a given kind
func QueryResources(ctx context.Context, kind string) *ent.ResourceQuery {
	ensureEntClient()
	return entClient.Resource.Query().
		Where(entresource.KindEQ(kind))
}

// QueryResourcesByLabels queries resources by kind and exact-match labels
func QueryResourcesByLabels(ctx context.Context, kind string, labels map[string]string) (*ent.ResourceQuery, error) {
	ensureEntClient()
	q := entClient.Resource.Query().Where(entresource.KindEQ(kind))
	for k, v := range labels {
		q = q.Where(entresource.HasLabelsWith(
			label.KeyEQ(k),
			label.ValueEQ(v),
		))
	}
	return q, nil
}

// Querydeviceprofiles returns a query builder for deviceprofiles
func Querydeviceprofiles(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "DeviceProfile")
}

// GetDeviceProfileByUID loads a single DeviceProfile by UID
func GetDeviceProfileByUID(ctx context.Context, uid string) (*v1.DeviceProfile, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("DeviceProfile")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load DeviceProfile %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.DeviceProfile), nil
}

// ListdeviceprofilesByLabels returns deviceprofiles matching all provided labels
func ListdeviceprofilesByLabels(ctx context.Context, labels map[string]string) ([]*v1.DeviceProfile, error) {
	q, err := QueryResourcesByLabels(ctx, "DeviceProfile", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.DeviceProfile, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.DeviceProfile))
	}
	return out, nil
}

// Queryfirmwaremanifests returns a query builder for firmwaremanifests
func Queryfirmwaremanifests(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "FirmwareManifest")
}

// GetFirmwareManifestByUID loads a single FirmwareManifest by UID
func GetFirmwareManifestByUID(ctx context.Context, uid string) (*v1.FirmwareManifest, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("FirmwareManifest")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load FirmwareManifest %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.FirmwareManifest), nil
}

// ListfirmwaremanifestsByLabels returns firmwaremanifests matching all provided labels
func ListfirmwaremanifestsByLabels(ctx context.Context, labels map[string]string) ([]*v1.FirmwareManifest, error) {
	q, err := QueryResourcesByLabels(ctx, "FirmwareManifest", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.FirmwareManifest, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.FirmwareManifest))
	}
	return out, nil
}

// Querylookupjobs returns a query builder for lookupjobs
func Querylookupjobs(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "LookupJob")
}

// GetLookupJobByUID loads a single LookupJob by UID
func GetLookupJobByUID(ctx context.Context, uid string) (*v1.LookupJob, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("LookupJob")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load LookupJob %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.LookupJob), nil
}

// ListlookupjobsByLabels returns lookupjobs matching all provided labels
func ListlookupjobsByLabels(ctx context.Context, labels map[string]string) ([]*v1.LookupJob, error) {
	q, err := QueryResourcesByLabels(ctx, "LookupJob", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.LookupJob, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.LookupJob))
	}
	return out, nil
}

// Queryupdatejobs returns a query builder for updatejobs
func Queryupdatejobs(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "UpdateJob")
}

// GetUpdateJobByUID loads a single UpdateJob by UID
func GetUpdateJobByUID(ctx context.Context, uid string) (*v1.UpdateJob, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("UpdateJob")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load UpdateJob %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.UpdateJob), nil
}

// ListupdatejobsByLabels returns updatejobs matching all provided labels
func ListupdatejobsByLabels(ctx context.Context, labels map[string]string) ([]*v1.UpdateJob, error) {
	q, err := QueryResourcesByLabels(ctx, "UpdateJob", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.UpdateJob, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.UpdateJob))
	}
	return out, nil
}

// Queryupdateprofiles returns a query builder for updateprofiles
func Queryupdateprofiles(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "UpdateProfile")
}

// GetUpdateProfileByUID loads a single UpdateProfile by UID
func GetUpdateProfileByUID(ctx context.Context, uid string) (*v1.UpdateProfile, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("UpdateProfile")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load UpdateProfile %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.UpdateProfile), nil
}

// ListupdateprofilesByLabels returns updateprofiles matching all provided labels
func ListupdateprofilesByLabels(ctx context.Context, labels map[string]string) ([]*v1.UpdateProfile, error) {
	q, err := QueryResourcesByLabels(ctx, "UpdateProfile", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.UpdateProfile, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.UpdateProfile))
	}
	return out, nil
}
