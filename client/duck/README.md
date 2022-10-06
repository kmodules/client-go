# Duck Typing

This package implements a kubebuilder controller-runtime compatible `Client`, `Lister` and `Controller`. To learn about Duck Typing, read [Knative Duck Typing](https://github.com/knative/pkg/blob/main/apis/duck/ABOUT.md).


## Differences with Knative Implementation

- Uses kubebuilder controller-runtime
- Use Typed clients and Informer/Lister to watch instead of the dynamic client in kantive implementation. This allows users of this package to also watch the underlying types directly if needed using the same lister/watcher.
- Instead of using JSON to marshal api types to duck types, we depend on the `Duckify` converter method. This allows us to take advantage of duck typing even when the JSON format is not compatible.

## Examples

**api**
- https://github.com/tamalsaha/duckdemo/blob/master/apis/core/v1alpha1/mypod_types.go
- https://github.com/tamalsaha/duckdemo/blob/master/apis/core/v1alpha1/mypod_conversion.go

**controller**
- https://github.com/tamalsaha/duckdemo/blob/master/controllers/core/mypod_controller.go
