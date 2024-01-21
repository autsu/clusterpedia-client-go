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

package customclient

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/clusterpedia-io/client-go/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

var parameterScheme = runtime.NewScheme()
var parameterCodec = runtime.NewParameterCodec(parameterScheme)
var versionV1 = schema.GroupVersion{Version: "v1"}

var _ Interface = &restClient{}
var _ ResourceInterface = &restResourceClient{}

func ConfigFor(inConfig *rest.Config) *rest.Config {
	config := rest.CopyConfig(inConfig)
	config.AcceptContentTypes = "application/json"
	config.ContentType = "application/json"
	config.NegotiatedSerializer = basicNegotiatedSerializer{}
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	return config
}

func NewForConfig(inConfig *rest.Config) (Interface, error) {
	config, err := client.ConfigFor(inConfig)
	if err != nil {
		return nil, err
	}

	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(config, httpClient)
}

func NewForConfigAndClient(inConfig *rest.Config, h *http.Client) (Interface, error) {
	config := ConfigFor(inConfig)
	// for serializing the options
	config.GroupVersion = &schema.GroupVersion{}
	config.APIPath = "/if-you-see-this-search-for-the-break"

	rc, err := rest.RESTClientForConfigAndClient(config, h)
	if err != nil {
		return nil, err
	}
	return &restClient{client: rc}, nil
}

type restResourceClient struct {
	client    *restClient
	namespace string
	resource  schema.GroupVersionResource
	openDebug bool
}

type restClient struct {
	client    *rest.RESTClient
	openDebug bool
}

func (c *restClient) Debug() Interface {
	c.openDebug = true
	return c
}

func (c *restClient) Resource(resource schema.GroupVersionResource) NamespaceableResourceInterface {
	return &restResourceClient{client: c, resource: resource, openDebug: c.openDebug}
}

func (c *restResourceClient) Namespace(ns string) ResourceInterface {
	ret := *c
	ret.namespace = ns
	return &ret
}

func (c *restResourceClient) List(ctx context.Context, opts metav1.ListOptions, params map[string]string, obj runtime.Object) error {
	req := rest.NewRequest(c.client.client)
	req.AbsPath(c.makeURLSegments("")...).SpecificallyVersionedParams(&opts, parameterCodec, versionV1)
	for key, value := range params {
		req.Param(key, value)
	}

	if c.openDebug {
		unescape, _ := url.QueryUnescape(req.URL().String())
		slog.Debug("", slog.String("req.URL", unescape))
	}
	return req.Do(ctx).Into(obj)
}

func (c *restResourceClient) makeURLSegments(name string) []string {
	url := []string{}
	if len(c.resource.Group) == 0 {
		url = append(url, "api")
	} else {
		url = append(url, "apis", c.resource.Group)
	}
	url = append(url, c.resource.Version)

	if len(c.namespace) > 0 {
		url = append(url, "namespaces", c.namespace)
	}

	url = append(url, c.resource.Resource)

	if len(name) > 0 {
		url = append(url, name)
	}

	return url
}
