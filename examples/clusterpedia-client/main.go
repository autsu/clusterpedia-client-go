/*
Copyright 2021 clusterpedia Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/clusterpedia-io/client-go/clusterpediaclient"
	"github.com/clusterpedia-io/client-go/tools/builder"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		//AddSource: true,
		Level: slog.LevelDebug,
	})))

	restConfig, err := ctrl.GetConfig()
	if err != nil {
		panic(err)
	}
	cc, err := clusterpediaclient.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	collectionResource, err := cc.PediaClusterV1beta1().CollectionResource().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, item := range collectionResource.Items {
		//fmt.Printf("resource info: %v\n", item)
		_ = item
	}

	// build listOptions
	options := builder.ListOptionsBuilder().
		Namespaces("kube-system").
		Namespaces("defualt").
		Options()

	resources, err := cc.PediaClusterV1beta1().CollectionResource().Fetch(context.TODO(), "workloads", options, nil)
	if err != nil {
		panic(err)
	}

	for _, item := range resources.Items {

		//fmt.Printf("resource info: %v\n", string(item.Raw))
		_ = item
	}

	options = builder.ListOptionsBuilder().Namespaces(metav1.NamespaceDefault).Options()
	resources, err = cc.PediaClusterV1beta1().CollectionResource().Fetch(context.TODO(), "workloads", options, map[string]string{
		"clusters":     "k3s-2",
		"onlyMetadata": "true",
	})
	if err != nil {
		panic(err)
	}

	for _, item := range resources.Items {
		deploy := &appsv1.Deployment{}
		unstructured.UnstructuredJSONScheme.Decode(item.Raw, &schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		}, deploy)
		//slog.Debug("resource info", slog.String("info", string(item.Raw)))
		slog.Debug("deploy",
			slog.String("namespace/name", fmt.Sprintf("%v/%v", deploy.Namespace, deploy.Name)))
		//fmt.Printf("resource info: %v\n", string(item.Raw))
		_ = item
	}

	options = builder.ListOptionsBuilder().Namespaces(metav1.NamespaceDefault).Options()
	resources, err = cc.PediaClusterV1beta1().CollectionResource().Fetch(context.TODO(), "any", options, map[string]string{
		"onlyMetadata": "true",
		"groups":       "apps",
		//"resources":    "",
	})
	if err != nil {
		panic(err)
	}
	for _, item := range resources.Items {
		slog.Debug("resource info", slog.String("info", string(item.Raw)))
		//fmt.Printf("resource info: %v\n", string(item.Raw))
	}
}
