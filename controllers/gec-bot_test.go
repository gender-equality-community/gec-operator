//go:build types
// +build types

package controllers

import (
	"os"
	"reflect"
	"testing"

	deploymentv1alpha1 "github.com/gender-equality-community/gec-operator/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type unmarshaler interface {
	Unmarshal([]byte) error
}

func unmarshalFile(fn string, i interface{}) (err error) {
	f, err := os.ReadFile(fn)
	if err != nil {
		return
	}

	return yaml.UnmarshalStrict(f, i)
}

var (
	bot = &deploymentv1alpha1.Cluster{
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
					Version: "v0.0.1",
				},
			},
			Slacker: deploymentv1alpha1.Slacker{
				App: deploymentv1alpha1.App{
					Version: "v0.0.1",
				},
			},
			Config: deploymentv1alpha1.Config{
				RedisURL: "redis.example.com:6379",
			},
		},
	}
)

func TestBot_ServiceAccount(t *testing.T) {
	expect := new(corev1.ServiceAccount)

	err := unmarshalFile("testdata/bot-sa.yaml", expect)
	if err != nil {
		t.Fatal(err)
	}

	received := serviceAccount(bot, deploymentv1alpha1.ClusterBot, GecBotLabels(bot))

	if !reflect.DeepEqual(expect, received) {
		got, err := yaml.Marshal(received)
		if err != nil {
			t.Fatal(err)
		}

		t.Errorf("expected\n%s", got)
	}
}

func TestBot_ConfigMap(t *testing.T) {
	expect := new(corev1.ConfigMap)

	err := unmarshalFile("testdata/bot-cm.yaml", expect)
	if err != nil {
		t.Fatal(err)
	}

	received := configmap(bot, deploymentv1alpha1.ClusterBot, GecBotLabels(bot), map[string]string{"SOMETHING": "also"})

	if !reflect.DeepEqual(expect, received) {
		got, err := yaml.Marshal(received)
		if err != nil {
			t.Fatal(err)
		}

		t.Errorf("expected\n%s", got)
	}
}

func TestBot_Deployment(t *testing.T) {
	expect := new(appsv1.Deployment)

	err := unmarshalFile("testdata/bot-deployment.yaml", expect)
	if err != nil {
		t.Fatal(err)
	}

	received := deployment(bot, deploymentv1alpha1.ClusterBot, GecBotLabels(bot), GecBotSelectors(bot))
	received.Spec.Template.Spec.SecurityContext = nil

	if !cmp.Equal(expect.Spec, received.Spec) {
		t.Fatal(cmp.Diff(expect, received))
	}
}

func TestBot_PVC(t *testing.T) {
	expect := new(corev1.PersistentVolumeClaim)

	err := unmarshalFile("testdata/bot-pvc.yaml", expect)
	if err != nil {
		t.Fatal(err)
	}

	received := pvc(bot, deploymentv1alpha1.ClusterBot, GecBotLabels(bot))
	if !reflect.DeepEqual(expect, received) {
		got, err := yaml.Marshal(received)
		if err != nil {
			t.Fatal(err)
		}

		t.Errorf("expected\n%s", got)
	}
}
