package controllers

import (
	appv1alpha1 "github.com/gender-equality-community/gec-operator/api/v1alpha1"
)

var gecSlackerUpserters = []upserter{
	ServiceAccount,
	ConfigMap,
	Deployment,
}

func GecSlackerSelectors(app *appv1alpha1.Cluster) map[string]string {
	return map[string]string{
		"cluster": app.Name,
		"app":     "gec-slacker",
	}
}

func GecSlackerLabels(app *appv1alpha1.Cluster) map[string]string {
	l := GecSlackerSelectors(app)
	l["version"] = app.Spec.Slacker.Version

	return l
}
