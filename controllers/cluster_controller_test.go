package controllers

import (
	"testing"
)

func TestClusterReconciler_SetupWithManager(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}
	}()

	r := new(ClusterReconciler)

	r.SetupWithManager(nil)
}

func TestRedisHostname(t *testing.T) {
	for _, test := range []struct {
		in     string
		expect string
	}{
		{"localhost:6379", "localhost"},
		{"redis://localhost:6379", "localhost"},
		{"localhost", "localhost"},
		{"redis://localhost", "localhost"},
		{"redis://localhost/0", "localhost"},
		{"redis://localhost:6379/0", "localhost"},
	} {
		t.Run(test.in, func(t *testing.T) {
			received := redisHostname(test.in)
			if test.expect != received {
				t.Errorf("expected %q, received %q", test.expect, received)
			}
		})
	}
}
