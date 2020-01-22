package finalizer

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/skygeario/k8s-controller/util/slice"
)

func Ensure(client client.Client, ctx context.Context, obj runtime.Object, finalizer string) (added bool, err error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return false, err
	}
	if slice.ContainsString(accessor.GetFinalizers(), finalizer) {
		return false, nil
	}
	accessor.SetFinalizers(append(accessor.GetFinalizers(), finalizer))
	added = true
	err = client.Update(ctx, obj)
	return
}

func Remove(client client.Client, ctx context.Context, obj runtime.Object, finalizer string) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetFinalizers(slice.RemoveString(accessor.GetFinalizers(), finalizer))
	return client.Update(ctx, obj)
}
