package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/clusterpedia-io/client-go/clusterpediaclient"
	"github.com/clusterpedia-io/client-go/tools/builder"
)

var cc *clusterpediaclient.ClusterpediaClient

func Init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		//AddSource: true,
		Level: slog.LevelDebug,
	})))

	restConfig, err := ctrl.GetConfig()
	if err != nil {
		panic(err)
	}
	c, err := clusterpediaclient.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}
	cc = c
}

func TestListCollectionResource(t *testing.T) {
	Init()

	// https://kubernetes.docker.internal:6443/apis/clusterpedia.io/v1beta1/collectionresources
	collectionResource, err := cc.PediaClusterV1beta1().CollectionResource().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, item := range collectionResource.Items {
		slog.Debug("resource info",
			slog.String("name", item.Name),
			slog.Any("resourceTypes", item.ResourceTypes),
		)
	}
}

func TestListWorkloads(t *testing.T) {
	Init()

	// build listOptions
	// 只查询 default 这个 namespace 下的资源
	options := builder.ListOptionsBuilder().
		Namespaces(metav1.NamespaceDefault).
		Options()

	// 查询
	// https://kubernetes.docker.internal:6443/apis/clusterpedia.io/v1beta1/collectionresources/workloads?labelSelector=search.clusterpedia.io/namespaces=default
	resources, err := cc.PediaClusterV1beta1().CollectionResource().Fetch(context.TODO(), "workloads", options, nil)
	if err != nil {
		panic(err)
	}

	for _, item := range resources.Items {
		us := &unstructured.Unstructured{}
		if err := json.Unmarshal(item.Raw, us); err != nil {
			panic(err)
		}
		gvk := us.GroupVersionKind()
		switch gvk.Kind {
		case "Deployment":
			deploy := &appsv1.Deployment{}
			_, _, err := unstructured.UnstructuredJSONScheme.Decode(item.Raw,
				&schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				}, deploy)
			if err != nil {
				panic(err)
			}
			slog.Debug("resource info",
				slog.Any("kind", gvk.Kind),
				slog.Any("namespace/name", fmt.Sprintf("%v/%v", deploy.Namespace, deploy.Name)),
			)
		case "StatefulSet":

		case "DaemonSet":
			ds := &appsv1.DaemonSet{}
			_, _, err := unstructured.UnstructuredJSONScheme.Decode(item.Raw,
				&schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "DaemonSet",
				}, ds)
			if err != nil {
				panic(err)
			}
			slog.Debug("resource info",
				slog.Any("kind", gvk.Kind),
				slog.Any("namespace/name", fmt.Sprintf("%v/%v", ds.Namespace, ds.Name)),
			)
		}
	}
}

func TestListKubeResources(t *testing.T) {
	Init()

	// build listOptions
	// 只查询 default 这个 namespace 下的资源
	options := builder.ListOptionsBuilder().
		Namespaces(metav1.NamespaceDefault).
		Options()
	resources, err := cc.PediaClusterV1beta1().CollectionResource().Fetch(context.TODO(), "kuberesources", options, map[string]string{
		"clusters": "k3s-2",
	})
	if err != nil {
		panic(err)
	}
	for _, item := range resources.Items {
		us := &unstructured.Unstructured{}
		if err := json.Unmarshal(item.Raw, us); err != nil {
			panic(err)
		}
		slog.Debug("resource info",
			slog.String("kind", us.GetKind()),
			slog.String("namespace/name", fmt.Sprintf("%v/%v", us.GetNamespace(), us.GetName())),
		)
	}
}

func TestListAny(t *testing.T) {
	Init()

	options := builder.ListOptionsBuilder().
		Namespaces(metav1.NamespaceDefault).
		Limit(10).
		Clusters("k3s2").
		Options()

	resources, err := cc.PediaClusterV1beta1().Debug().CollectionResource().Fetch(context.TODO(), "any", options, map[string]string{
		"onlyMetadata": "true",
		// groups 指定一组资源的组和版本，多个组版本使用 , 分隔，组版本格式为 <group>/<version>，也可以不指定 version
		// 如果是 group 是 core，直接指定为空字符串即可
		// 比如下面就代表指定了 apps 和 core 这两个 group，没有指定 version
		"groups": "apps,,",
		// resources 可以指定具体的资源类型，多个资源类型使用 , 分隔，资源类型格式为 <group>/<version>/<resource>，
		// 也可以不指定版本 <group>/<resource>
		"resources": "apps/v1/deployments,apps/daemonsets,/pods",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range resources.Items {
		us := &unstructured.Unstructured{}
		if err := json.Unmarshal(item.Raw, us); err != nil {
			panic(err)
		}
		slog.Debug("resource info",
			slog.String("kind", us.GetKind()),
			slog.String("namespace/name", fmt.Sprintf("%v/%v", us.GetNamespace(), us.GetName())),
		)
	}
}

func LogNamespaceAndName(gvk schema.GroupVersionKind, obj metav1.Object) {
	slog.Debug("resource info",
		slog.Any("kind", gvk.Kind),
		slog.Any("namespace/name", fmt.Sprintf("%v/%v", obj.GetNamespace(), obj.GetName())),
	)
}
