package controllers

import (
	appv1alpha1 "github.com/gender-equality-community/gec-operator/api/v1alpha1"
)

var gecProcessorUpserters = []upserter{
	ServiceAccount,
	ConfigMap,
	Deployment,
}

func GecProcessorSelectors(app *appv1alpha1.Cluster) map[string]string {
	return map[string]string{
		"cluster": app.Name,
		"app":     "gec-processor",
	}
}

func GecProcessorLabels(app *appv1alpha1.Cluster) map[string]string {
	l := GecProcessorSelectors(app)
	l["version"] = app.Spec.Processor.Version
	l["sbom"] = app.Spec.Processor.SBOM()

	return l
}
