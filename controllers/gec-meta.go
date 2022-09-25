package controllers

import (
	appv1alpha1 "github.com/gender-equality-community/gec-operator/api/v1alpha1"
)

func GecMetaSelectors(app *appv1alpha1.Cluster) map[string]string {
	return map[string]string{
		"cluster": app.Name,
		"app":     "metadata",
	}
}

func GecMetaLabels(app *appv1alpha1.Cluster) map[string]string {
	return GecMetaSelectors(app)
}
