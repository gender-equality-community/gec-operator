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

package v1alpha1

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type App struct {
	// See: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	// we prefix 'v' to the version too, since that's what we slap on the front of our git and container tags.
	// +kubebuilder:validation:Pattern=`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	Version string `json:"version"`
}

type Bot struct {
	App `json:",inline"`
}

func (b Bot) Image() string {
	return taggedImage(botContainerImage, b.Version)
}

func (b Bot) HasValidSignature() bool {
	// skipped... for now
	return true
}

func (b Bot) SBOM() string {
	return sbomURL(botContainerImage, b.Version)
}

type Processor struct {
	App `json:",inline"`
}

func (p Processor) Image() string {
	return taggedImage(processorContainerImage, p.Version)
}

func (p Processor) HasValidSignature() bool {
	// skipped... for now
	return true
}

func (p Processor) SBOM() string {
	return sbomURL(processorContainerImage, p.Version)
}

type Slacker struct {
	App `json:",inline"`
}

func (s Slacker) Image() string {
	return taggedImage(slackerContainerImage, s.Version)
}

func (s Slacker) HasValidSignature() bool {
	// skipped... for now
	return true
}

func (s Slacker) SBOM() string {
	return sbomURL(slackerContainerImage, s.Version)
}

type Config struct {
	RedisURL string `json:"redis_url"`
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	Bot       Bot       `json:"bot"`
	Processor Processor `json:"processor"`
	Slacker   Slacker   `json:"slacker"`
	Config    Config    `json:"config"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

func (c Cluster) InClusterName(ca ClusterApp) string {
	return fmt.Sprintf("%s_%s", c.Name, ca.String())
}

func (c Cluster) InClusterImage(ca ClusterApp) string {
	switch ca {
	case ClusterBot:
		return c.Spec.Bot.Image()

	case ClusterProcessor:
		return c.Spec.Processor.Image()

	case ClusterSlacker:
		return c.Spec.Slacker.Image()

	default:
		return ""
	}
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}

func taggedImage(image, tag string) string {
	return fmt.Sprintf("%s:%s", image, tag)
}

func sbomURL(image, tag string) string {
	imageS := strings.Split(image, "/")
	repo := imageS[len(imageS)-1]

	return fmt.Sprintf("https://github.com/gender-equality-community/%s/releases/download/%s/bom.json",
		repo,
		tag,
	)
}
