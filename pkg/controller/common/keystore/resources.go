// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package keystore

import (
	"strings"

	commonv1alpha1 "github.com/elastic/cloud-on-k8s/pkg/apis/common/v1alpha1"
	"github.com/elastic/cloud-on-k8s/pkg/controller/common/driver"
	"github.com/elastic/cloud-on-k8s/pkg/controller/common/name"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("keystore")

// Resources holds all the resources needed to create a keystore in Kibana or in the APM server.
type Resources struct {
	// volume which contains the keystore data as provided by the user
	Volume corev1.Volume
	// init container used to create the keystore
	InitContainer corev1.Container
	// version of the secret provided by the user
	Version string
}

// HasKeystore interface represents an Elastic Stack application that offers a keystore which in ECK
// is populated using a user-provided secret containing secure settings.
type HasKeystore interface {
	metav1.Object
	runtime.Object
	SecureSettings() []commonv1alpha1.SecretSource
	// Kind can technically be retrieved from metav1.Object, but there is a bug preventing us to retrieve it
	// see https://github.com/kubernetes-sigs/controller-runtime/issues/406
	Kind() string
}

// NewResources optionally returns a volume and init container to include in pods,
// in order to create a Keystore from a Secret containing secure settings provided by
// the user and referenced in the Elastic Stack application spec.
func NewResources(
	r driver.Interface,
	hasKeystore HasKeystore,
	namer name.Namer,
	labels map[string]string,
	initContainerParams InitContainerParameters,
) (*Resources, error) {
	// setup a volume from the user-provided secure settings secret
	secretVolume, version, err := secureSettingsVolume(r, hasKeystore, labels, namer)
	if err != nil {
		return nil, err
	}
	if secretVolume == nil {
		// nothing to do
		return nil, nil
	}

	// build an init container to create the keystore from the secure settings volume
	initContainer, err := initContainer(
		*secretVolume,
		strings.ToLower(hasKeystore.Kind()),
		initContainerParams,
	)
	if err != nil {
		return nil, err
	}

	return &Resources{
		Volume:        secretVolume.Volume(),
		InitContainer: initContainer,
		Version:       version,
	}, nil
}
