package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	cluster = Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testing",
		},
		Spec: ClusterSpec{
			Bot:       Bot{App: App{Version: "v0.1.0"}},
			Processor: Processor{App: App{Version: "v0.1.0"}},
			Slacker:   Slacker{App: App{Version: "v0.1.0"}},
		},
	}
)

type sbommer interface {
	SBOM() string
}

func TestCluster_InClusterName(t *testing.T) {
	for _, test := range []struct {
		ca     ClusterApp
		expect string
	}{
		{ClusterBot, "testing-gec-bot"},
		{ClusterProcessor, "testing-gec-processor"},
		{ClusterSlacker, "testing-gec-slacker"},
	} {
		t.Run(test.ca.String(), func(t *testing.T) {
			received := cluster.InClusterName(test.ca)

			if test.expect != received {
				t.Errorf("expected %q, received %q", test.expect, received)
			}
		})
	}
}

func TestCluster_InClusterImage(t *testing.T) {
	for _, test := range []struct {
		ca     ClusterApp
		expect string
	}{
		{ClusterBot, "ghcr.io/gender-equality-community/gec-bot:v0.1.0"},
		{ClusterProcessor, "ghcr.io/gender-equality-community/gec-processor:v0.1.0"},
		{ClusterSlacker, "ghcr.io/gender-equality-community/gec-slacker:v0.1.0"},
		{UnknownClusterApp, ""},
	} {
		t.Run(test.ca.String(), func(t *testing.T) {
			received := cluster.InClusterImage(test.ca)

			if test.expect != received {
				t.Errorf("expected %q, received %q", test.expect, received)
			}
		})
	}
}

func TestHasValidSignature(t *testing.T) {
	for _, app := range []bool{
		cluster.Spec.Bot.HasValidSignature(),
		cluster.Spec.Processor.HasValidSignature(),
		cluster.Spec.Slacker.HasValidSignature(),
	} {
		t.Run("", func(t *testing.T) {
			if app != true {
				t.Error("expected true")
			}
		})
	}
}

func TestSBOM(t *testing.T) {
	for _, test := range []struct {
		ca     ClusterApp
		f      sbommer
		expect string
	}{
		{ClusterBot, cluster.Spec.Bot, "https://github.com/gender-equality-community/gec-bot/releases/download/v0.1.0/bom.json"},
		{ClusterProcessor, cluster.Spec.Processor, "https://github.com/gender-equality-community/gec-processor/releases/download/v0.1.0/bom.json"},
		{ClusterSlacker, cluster.Spec.Slacker, "https://github.com/gender-equality-community/gec-slacker/releases/download/v0.1.0/bom.json"},
	} {
		t.Run(test.ca.String(), func(t *testing.T) {
			received := test.f.SBOM()

			if test.expect != received {
				t.Errorf("expected %q, received %q", test.expect, received)
			}
		})
	}
}
