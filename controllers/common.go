package controllers

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	deploymentv1alpha1 "github.com/gender-equality-community/gec-operator/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Project string
)

type upserter func(context.Context, client.Client, *runtime.Scheme, *deploymentv1alpha1.Cluster, deploymentv1alpha1.ClusterApp, map[string]string, map[string]string) (time.Duration, error)

func init() {
	var err error

	Project, err = getenv("PROJECT")
	if err != nil {
		panic(err)
	}
}

func getenv(v string) (s string, err error) {
	s, ok := os.LookupEnv(v)
	if !ok || s == "" {
		err = fmt.Errorf("%s environment variable not set or empty", v)
	}

	return
}

func gcpServiceAccount(name string) string {
	return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, Project)
}

func ServiceAccount(ctx context.Context, c client.Client, s *runtime.Scheme, app *deploymentv1alpha1.Cluster, ca deploymentv1alpha1.ClusterApp, labels, selectors map[string]string) (requeue time.Duration, err error) {
	sa := serviceAccount(app, ca, labels)

	err = ctrl.SetControllerReference(app, sa, s)
	if err != nil {
		return
	}

	found := &corev1.ServiceAccount{}

	err = c.Get(ctx, types.NamespacedName{Name: sa.Name, Namespace: app.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(ctx, sa)
		if err != nil {
			return
		}

		return
	}

	if !reflect.DeepEqual(found.ImagePullSecrets, sa.ImagePullSecrets) {
		diff := cmp.Diff(found.ImagePullSecrets, sa.ImagePullSecrets)
		fmt.Println(diff)

		err = c.Update(ctx, sa)
		if err == nil {
			requeue = time.Second
		}
	}

	return
}

func serviceAccount(app *deploymentv1alpha1.Cluster, ca deploymentv1alpha1.ClusterApp, labels map[string]string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.InClusterName(ca),
			Namespace: app.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"iam.gke.io/gcp-service-account": gcpServiceAccount(app.Name),
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{Name: "ecr-pull"},
		},
	}
}

func ConfigMap(ctx context.Context, c client.Client, s *runtime.Scheme, app *deploymentv1alpha1.Cluster, ca deploymentv1alpha1.ClusterApp, labels, selectors map[string]string) (requeue time.Duration, err error) {
	cm := configmap(app, ca, labels, ctx.Value("config").(map[string]string))

	err = ctrl.SetControllerReference(app, cm, s)
	if err != nil {
		return
	}

	found := &corev1.ConfigMap{}

	err = c.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: app.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(ctx, cm)
		if err != nil {
			return
		}

		return
	}

	if !reflect.DeepEqual(found.Data, cm.Data) {
		diff := cmp.Diff(found.Data, cm.Data)
		fmt.Println(diff)

		err = c.Update(ctx, cm)
		if err == nil {
			requeue = time.Second
		}
	}

	return
}

func configmap(app *deploymentv1alpha1.Cluster, ca deploymentv1alpha1.ClusterApp, labels, config map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.InClusterName(ca),
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Data: config,
	}
}

func Deployment(ctx context.Context, c client.Client, s *runtime.Scheme, app *deploymentv1alpha1.Cluster, ca deploymentv1alpha1.ClusterApp, labels, selectors map[string]string) (requeue time.Duration, err error) {
	d := deployment(app, ca, labels, selectors)

	err = ctrl.SetControllerReference(app, d, s)
	if err != nil {
		return
	}

	found := &appsv1.Deployment{}

	err = c.Get(ctx, types.NamespacedName{Name: d.Name, Namespace: app.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(ctx, d)
		if err != nil {
			return
		}

		return
	}

	if !reflect.DeepEqual(found.Spec.Template.Spec, d.Spec.Template.Spec) {
		diff := cmp.Diff(found.Spec.Template.Spec, d.Spec.Template.Spec)
		fmt.Println(diff)

		err = c.Update(ctx, d)
		if err == nil {
			requeue = time.Second
		}
	}

	return
}

func deployment(app *deploymentv1alpha1.Cluster, ca deploymentv1alpha1.ClusterApp, labels, selectors map[string]string) *appsv1.Deployment {
	var (
		replicas           int32 = 1
		optional                 = true
		enableServiceLinks       = false
		automountSAToken         = false
		terminationGrace   int64 = 30
		trueVal                  = true
		falseVal                 = false
	)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectors,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: app.Name,
					Containers: []corev1.Container{{
						Image: app.InClusterImage(ca),
						Name:  app.InClusterName(ca),
						Resources: corev1.ResourceRequirements{
							Limits:   ca.Resources(),
							Requests: ca.Resources(),
						},
						VolumeMounts: ca.VolumeMount(app.InClusterName(ca)),
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: app.InClusterName(ca),
									},
								},
							},
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("%s-override", app.InClusterName(ca)),
									},
									Optional: &optional,
								},
							},
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("%s-override", app.InClusterName(ca)),
									},
									Optional: &optional,
								},
							},
						},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
						TTY:                      true,
						ImagePullPolicy:          corev1.PullIfNotPresent,
						SecurityContext: &corev1.SecurityContext{
							ReadOnlyRootFilesystem: &trueVal,
							Privileged:             &falseVal,
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{"ALL"},
							},
							AllowPrivilegeEscalation: &falseVal,
							RunAsNonRoot:             &trueVal,
							SeccompProfile: &corev1.SeccompProfile{
								Type: corev1.SeccompProfileTypeRuntimeDefault,
							},
						},
					}},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &terminationGrace,
					DNSPolicy:                     corev1.DNSClusterFirst,
					DeprecatedServiceAccount:      app.Name,
					SecurityContext:               &corev1.PodSecurityContext{},
					SchedulerName:                 "default-scheduler",
					Volumes:                       ca.Volume(app.InClusterName(ca)),
					EnableServiceLinks:            &enableServiceLinks,
					AutomountServiceAccountToken:  &automountSAToken,
				},
			},
		},
	}
}

func PVC(ctx context.Context, c client.Client, s *runtime.Scheme, app *deploymentv1alpha1.Cluster, ca deploymentv1alpha1.ClusterApp, labels, selectors map[string]string) (requeue time.Duration, err error) {
	p := pvc(app, ca, labels)

	err = ctrl.SetControllerReference(app, p, s)
	if err != nil {
		return
	}

	found := &corev1.PersistentVolumeClaim{}

	err = c.Get(ctx, types.NamespacedName{Name: p.Name, Namespace: app.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = c.Create(ctx, p)
	}

	return
}

func pvc(app *deploymentv1alpha1.Cluster, ca deploymentv1alpha1.ClusterApp, labels map[string]string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.InClusterName(ca),
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: deploymentv1alpha1.VolumeSize,
				},
			},
		},
	}
}
