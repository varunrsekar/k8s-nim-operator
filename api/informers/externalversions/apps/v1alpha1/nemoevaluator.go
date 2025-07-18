/*
Copyright 2025.

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
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	appsv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	internalinterfaces "github.com/NVIDIA/k8s-nim-operator/api/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/listers/apps/v1alpha1"
	versioned "github.com/NVIDIA/k8s-nim-operator/api/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NemoEvaluatorInformer provides access to a shared informer and lister for
// NemoEvaluators.
type NemoEvaluatorInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.NemoEvaluatorLister
}

type nemoEvaluatorInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNemoEvaluatorInformer constructs a new informer for NemoEvaluator type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNemoEvaluatorInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNemoEvaluatorInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNemoEvaluatorInformer constructs a new informer for NemoEvaluator type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNemoEvaluatorInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1alpha1().NemoEvaluators(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1alpha1().NemoEvaluators(namespace).Watch(context.TODO(), options)
			},
		},
		&appsv1alpha1.NemoEvaluator{},
		resyncPeriod,
		indexers,
	)
}

func (f *nemoEvaluatorInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNemoEvaluatorInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *nemoEvaluatorInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&appsv1alpha1.NemoEvaluator{}, f.defaultInformer)
}

func (f *nemoEvaluatorInformer) Lister() v1alpha1.NemoEvaluatorLister {
	return v1alpha1.NewNemoEvaluatorLister(f.Informer().GetIndexer())
}
