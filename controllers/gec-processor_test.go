//go:build types
// +build types

package controllers

import (
	"reflect"
	"testing"

	deploymentv1alpha1 "github.com/gender-equality-community/gec-operator/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var (
	processor = &deploymentv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test-cluster",
			Namespace: "testing",
		},
		Spec: deploymentv1alpha1.ClusterSpec{
			Bot: deploymentv1alpha1.Bot{
				App: deploymentv1alpha1.App{
					Version: "v0.0.1",
				},
			},
			Processor: deploymentv1alpha1.Processor{
				App: deploymentv1alpha1.App{
					Version: "v0.0.2",
				},
			},
			Slacker: deploymentv1alpha1.Slacker{
				App: deploymentv1alpha1.App{
					Version: "v0.0.3",
				},
			},
			Config: deploymentv1alpha1.Config{
				RedisURL: "redis.example.com:6379",
			},
		},
	}
)

func TestProcessor_ServiceAccount(t *testing.T) {
	expect := new(corev1.ServiceAccount)

	err := unmarshalFile("testdata/processor-sa.yaml", expect)
	if err != nil {
		t.Fatal(err)
	}

	received := serviceAccount(processor, deploymentv1alpha1.ClusterProcessor, GecProcessorLabels(processor))

	if !reflect.DeepEqual(expect, received) {
		got, err := yaml.Marshal(received)
		if err != nil {
			t.Fatal(err)
		}

		t.Errorf("expected\n%s", got)
	}
}

func TestProcessor_ConfigMap(t *testing.T) {
	expect := new(corev1.ConfigMap)

	err := unmarshalFile("testdata/processor-cm.yaml", expect)
	if err != nil {
		t.Fatal(err)
	}

	received := configmap(processor, deploymentv1alpha1.ClusterProcessor, GecProcessorLabels(processor), map[string]string{"SOMETHING": "also"})

	if !reflect.DeepEqual(expect, received) {
		got, err := yaml.Marshal(received)
		if err != nil {
			t.Fatal(err)
		}

		t.Errorf("expected\n%s", got)
	}
}

func TestProcessor_Deployment(t *testing.T) {
	expect := new(appsv1.Deployment)

	err := unmarshalFile("testdata/processor-deployment.yaml", expect)
	if err != nil {
		t.Fatal(err)
	}

	received := deployment(processor, deploymentv1alpha1.ClusterProcessor, GecProcessorLabels(processor), GecProcessorSelectors(processor))
	received.Spec.Template.Spec.SecurityContext = nil

	if !cmp.Equal(expect.Spec, received.Spec) {
		t.Fatal(cmp.Diff(expect, received))
	}
}
