package util

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
  "github.com/mitchellh/mapstructure"
	// "k8s.io/apimachinery/pkg/runtime"
)

func MapUnstructuredToStruct(obj *unstructured.Unstructured, target any) error {
  mapstructure.Decode(obj.Object, &target)

  return nil
}
