/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/gender-equality-community/gec-operator/api/v1alpha1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.gec,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.gec,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.gec,resources=clusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	app := new(appv1alpha1.Cluster)
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("No consumer object found, probably just been deleted /shrug")

			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get API")

		return ctrl.Result{}, err
	}

	requeue, err := r.Upsert(ctx, gecBotUpserters, appv1alpha1.ClusterBot, app, GecBotLabels(app), GecBotSelectors(app), map[string]string{"REDIS_ADDR": app.Spec.Config.RedisURL, "DATABASE": "/database/bot.db"})
	if err != nil || requeue > 0 {
		return ctrl.Result{RequeueAfter: requeue}, err
	}

	requeue, err = r.Upsert(ctx, gecProcessorUpserters, appv1alpha1.ClusterProcessor, app, GecProcessorLabels(app), GecProcessorSelectors(app), map[string]string{"REDIS_HOSTNAME": redisHostname(app.Spec.Config.RedisURL)})
	if err != nil || requeue > 0 {
		return ctrl.Result{RequeueAfter: requeue}, err
	}

	requeue, err = r.Upsert(ctx, gecSlackerUpserters, appv1alpha1.ClusterSlacker, app, GecSlackerLabels(app), GecSlackerSelectors(app), map[string]string{"REDIS_ADDR": app.Spec.Config.RedisURL, "INCOMING_STREAM": "gec-processed", "OUTGOING_STREAM": "gec-responses"})

	// Write final status
	ctx = context.WithValue(ctx, "config", map[string]string{
		"bot_sbom":       app.Spec.Bot.SBOM(),
		"processor_sbom": app.Spec.Processor.SBOM(),
		"slacker_bom":    app.Spec.Slacker.SBOM(),
		"last_deploy":    time.Now().String(),
		"version":        Version,
		"project":        Project,
	})
	requeue, err = ConfigMap(ctx, r.Client, r.Scheme, app, appv1alpha1.ClusterMeta, GecMetaLabels(app), nil)

	return ctrl.Result{RequeueAfter: requeue}, err
}

func (r *ClusterReconciler) Upsert(ctx context.Context, upserters []upserter, ca appv1alpha1.ClusterApp, app *appv1alpha1.Cluster, labels, selectors, config map[string]string) (requeue time.Duration, err error) {
	ctx = context.WithValue(ctx, "config", config)
	for _, f := range upserters {
		requeue, err = f(ctx, r.Client, r.Scheme, app, ca, labels, selectors)
		if err != nil || requeue > 0 {
			return
		}
	}

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Cluster{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}

func redisHostname(s string) string {
	if !isUri(s) {
		s = fmt.Sprintf("redis://%s", s)
	}

	// #nosec
	u, _ := url.Parse(s)

	out, _, err := net.SplitHostPort(u.Host)
	if err != nil && err.(*net.AddrError).Err == "missing port in address" {
		return u.Host
	}

	return out
}

func isUri(s string) bool {
	return strings.Contains(s, "://")
}
