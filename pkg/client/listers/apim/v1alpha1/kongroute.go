/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/cdiscount/kong-operator/pkg/apis/apim/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// KongRouteLister helps list KongRoutes.
type KongRouteLister interface {
	// List lists all KongRoutes in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.KongRoute, err error)
	// KongRoutes returns an object that can list and get KongRoutes.
	KongRoutes(namespace string) KongRouteNamespaceLister
	KongRouteListerExpansion
}

// kongRouteLister implements the KongRouteLister interface.
type kongRouteLister struct {
	indexer cache.Indexer
}

// NewKongRouteLister returns a new KongRouteLister.
func NewKongRouteLister(indexer cache.Indexer) KongRouteLister {
	return &kongRouteLister{indexer: indexer}
}

// List lists all KongRoutes in the indexer.
func (s *kongRouteLister) List(selector labels.Selector) (ret []*v1alpha1.KongRoute, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.KongRoute))
	})
	return ret, err
}

// KongRoutes returns an object that can list and get KongRoutes.
func (s *kongRouteLister) KongRoutes(namespace string) KongRouteNamespaceLister {
	return kongRouteNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// KongRouteNamespaceLister helps list and get KongRoutes.
type KongRouteNamespaceLister interface {
	// List lists all KongRoutes in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.KongRoute, err error)
	// Get retrieves the KongRoute from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.KongRoute, error)
	KongRouteNamespaceListerExpansion
}

// kongRouteNamespaceLister implements the KongRouteNamespaceLister
// interface.
type kongRouteNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all KongRoutes in the indexer for a given namespace.
func (s kongRouteNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.KongRoute, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.KongRoute))
	})
	return ret, err
}

// Get retrieves the KongRoute from the indexer for a given namespace and name.
func (s kongRouteNamespaceLister) Get(name string) (*v1alpha1.KongRoute, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("kongroute"), name)
	}
	return obj.(*v1alpha1.KongRoute), nil
}
