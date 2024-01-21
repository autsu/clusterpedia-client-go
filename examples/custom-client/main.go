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
	"log"
	"log/slog"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/clusterpedia-io/client-go/customclient"
	"github.com/clusterpedia-io/client-go/tools/builder"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})))

	config, err := ctrl.GetConfig()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}
	customClient, err := customclient.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to init customClient: %v", err)
	}

	// ================== /apis/apps/v1/deployments
	deploys := &appsv1.DeploymentList{}
	options := builder.ListOptionsBuilder().
		Offset(0).Limit(5).
		RemainingCount(). // 返回剩余资源数量，配合分页使用
		OrderBy("created_at", true).
		OrderBy("namespace", false).
		Options()

	// https://kubernetes.docker.internal:6443/apis/clusterpedia.io/v1beta1/resources/apis/apps/v1/namespaces/default/deployments?clusters=k3s-1&continue=0&labelSelector=search.clusterpedia.io/with-remaining-count=true&limit=10&timeout=10s
	if err := customClient.
		Resource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}).
		Namespace(metav1.NamespaceDefault).
		List(context.TODO(), options, nil, deploys); err != nil {
		panic(err)
	}

	for _, item := range deploys.Items {
		fmt.Printf("namespace: %s, name: %s\n", item.Namespace, item.Name)
	}

	fmt.Printf("\n\n\n")

	// =================== /api/v1/pods

	options = builder.ListOptionsBuilder().
		Offset(0).Limit(5).
		RemainingCount(). // 返回剩余资源数量，配合分页使用
		OrderBy("created_at", true).
		OrderBy("namespace", false).
		Clusters("k3s2").
		FieldSelector("status.phase", []string{"Running"}).
		Options()

	pods := &corev1.PodList{}
	if err := customClient.
		Debug().
		Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}).
		Namespace(metav1.NamespaceDefault).
		List(context.TODO(), options, map[string]string{"onlyMetadata": "true"}, pods); err != nil {
		panic(err)
	}
	if pods.RemainingItemCount != nil {
		fmt.Println("剩余资源数量：", *pods.RemainingItemCount)
	}
	for _, item := range pods.Items {
		fmt.Printf("namespace: %s, name: %s\n", item.Namespace, item.Name)
		//fmt.Printf("%+v\n", item)
	}

	fmt.Printf("\n\n\n")

	// 根据名字查询
	options = builder.ListOptionsBuilder().Names("nginx1").Options()
	if err := customClient.
		Debug().
		Resource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}).
		List(context.TODO(), options, nil, deploys); err != nil {
		panic(err)
	}
	for _, item := range deploys.Items {
		fmt.Printf("namespace: %s, name: %s\n", item.Namespace, item.Name)
	}

	fmt.Printf("\n\n\n")

	// 根据名字模糊查询
	options = builder.ListOptionsBuilder().FuzzyNames("ngin").Options()
	if err := customClient.
		Debug().
		Resource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}).
		List(context.TODO(), options, nil, deploys); err != nil {
		panic(err)
	}
	for _, item := range deploys.Items {
		fmt.Printf("namespace: %s, name: %s\n", item.Namespace, item.Name)
	}
}
