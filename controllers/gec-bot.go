package controllers

import (
	appv1alpha1 "github.com/gender-equality-community/gec-operator/api/v1alpha1"
)

var gecBotUpserters = []upserter{
	ServiceAccount,
	ConfigMap,
	PVC,
	Deployment,
}

func GecBotSelectors(app *appv1alpha1.Cluster) map[string]string {
	return map[string]string{
		"cluster": app.Name,
		"app":     "gec-bot",
	}
}

func GecBotLabels(app *appv1alpha1.Cluster) map[string]string {
	l := GecBotSelectors(app)
	l["version"] = app.Spec.Bot.Version

	return l
}
