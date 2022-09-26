package v1alpha1

import (
	"testing"
)

func TestClusterApp_String(t *testing.T) {
	for _, test := range []struct {
		ca     ClusterApp
		expect string
	}{
		{ClusterBot, "gec-bot"},
		{ClusterProcessor, "gec-processor"},
		{ClusterSlacker, "gec-slacker"},
		{ClusterMeta, "meta"},
		{UnknownClusterApp, "unknown"},
		{ClusterApp(240), "unknown"},
	} {
		t.Run(test.expect, func(t *testing.T) {
			received := test.ca.String()
			if test.expect != received {
				t.Errorf("expected %q, received %q", test.expect, received)
			}
		})
	}
}

func TestClusterApp_Resources(t *testing.T) {
	for _, test := range []struct {
		ca        ClusterApp
		expectCPU string
		expectMem string
	}{
		{ClusterBot, "100m", "64Mi"},
		{ClusterProcessor, "200m", "128Mi"},
	} {
		t.Run(test.ca.String(), func(t *testing.T) {
			received := test.ca.Resources()

			t.Run("CPU", func(t *testing.T) {
				if test.expectCPU != received.Cpu().String() {
					t.Errorf("expected %q, received %q", test.expectCPU, received.Cpu().String())
				}
			})

			t.Run("memory", func(t *testing.T) {
				if test.expectMem != received.Memory().String() {
					t.Errorf("expected %q, received %q", test.expectMem, received.Memory().String())
				}
			})

		})
	}
}

func TestClusterApp_VolumeMount(t *testing.T) {
	for _, test := range []struct {
		ca         ClusterApp
		expectPath string
	}{
		{ClusterBot, "/database/"},
		{ClusterProcessor, ""},
	} {
		t.Run(test.ca.String(), func(t *testing.T) {
			received := test.ca.VolumeMount("db")
			if len(received) == 0 && test.expectPath != "" {
				t.Error("expected volume mount, received none")
			} else if len(received) == 1 {
				if test.expectPath == "" {
					t.Error("unexpected volume mount")
				} else if test.expectPath != received[0].MountPath {
					t.Errorf("expected %q, received %q", test.expectPath, received[0].MountPath)
				}

			} else if len(received) > 1 {
				t.Errorf("unexpected volume mounts %#v", received)
			}
		})
	}
}

func TestClusterApp_Volume(t *testing.T) {
	for _, test := range []struct {
		ca        ClusterApp
		expectLen int
	}{
		{ClusterBot, 1},
		{ClusterProcessor, 0},
	} {
		t.Run(test.ca.String(), func(t *testing.T) {
			received := len(test.ca.Volume("foo"))
			if test.expectLen != received {
				t.Errorf("expected %d volume(s), received %d", test.expectLen, received)
			}
		})
	}
}
