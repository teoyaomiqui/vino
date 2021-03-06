/*


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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VinoSpec defines the desired state of Vino
type VinoSpec struct {
	// Define nodelabel parameters
	NodeSelector *NodeSelector `json:"nodelabels,omitempty"`
	// Define CPU configuration
	CPUConfiguration *CPUConfiguration `json:"configuration,omitempty"`
	// Define network Parametes
	Network *Network `json:"networks,omitempty"`
	// Define node details
	Node []NodeSet `json:"nodes,omitempty"`
	// DaemonSetOptions defines how vino will spawn daemonset on nodes
	DaemonSetOptions DaemonSetOptions `json:"daemonSetOptions,omitempty"`
}

// NodeSelector identifies nodes to create VMs on
type NodeSelector struct {
	// Node type needs to specified
	MatchLabels map[string]string `json:"matchLabels"`
}

// CPUConfiguration CPU node configuration
type CPUConfiguration struct {
	//Exclude CPU example 0-4,54-60
	CPUExclude string `json:"cpuExclude,omitempty"`
}

// Network defines libvirt networks
type Network struct {
	//Network Parameter defined
	Name            string    `json:"name,omitempty"`
	SubNet          string    `json:"subnet,omitempty"`
	AllocationStart string    `json:"allocationStart,omitempty"`
	AllocationStop  string    `json:"allocationStop,omitempty"`
	DNSServers      []string  `json:"dns_servers,omitempty"`
	Routes          *VMRoutes `json:"routes,omitempty"`
}

// VMRoutes defined
type VMRoutes struct {
	To  string `json:"to,omitempty"`
	Via string `json:"via,omitempty"`
}

//NodeSet node definitions
type NodeSet struct {
	//Parameter for Node master or worker-standard
	Name                      string              `json:"name,omitempty"`
	Count                     int                 `json:"count,omitempty"`
	NodeLabel                 *VMNodeFlavor       `json:"labels,omitempty"`
	LibvirtTemplateDefinition NamespacedName      `json:"libvirtTemplate,omitempty"`
	NetworkInterface          *NetworkInterface   `json:"networkInterfaces,omitempty"`
	DiskDrives                *DiskDrivesTemplate `json:"diskDrives,omitempty"`
}

// VMNodeFlavor labels for node to be annotated
type VMNodeFlavor struct {
	VMFlavor map[string]string `json:"vmFlavor,omitempty"`
}

// NamespacedName to be used to spawn VMs
type NamespacedName struct {
	Name      string `json:"Name,omitempty"`
	Namespace string `json:"Namespace,omitempty"`
}

// DaemonSetOptions be used to spawn vino-builder, libvrirt, sushy an
type DaemonSetOptions struct {
	Template     NamespacedName `json:"namespacedName,omitempty"`
	LibvirtImage string         `json:"libvirtImage,omitempty"`
	SushyImage   string         `json:"sushyImage,omitempty"`
	VinoBuilder  string         `json:"vinoBuilderImage,omitempty"`
	NodeLabeler  string         `json:"nodeAnnotatorImage,omitempty"`
}

// NetworkInterface define interface on the VM
type NetworkInterface struct {
	//Define parameter for netwok interfaces
	Name        string            `json:"name,omitempty"`
	Type        string            `json:"type,omitempty"`
	NetworkName string            `json:"network,omitempty"`
	MTU         int               `json:"mtu,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}

// DiskDrivesTemplate defines diks on the VM
type DiskDrivesTemplate struct {
	Name    string       `json:"name,omitempty"`
	Type    string       `json:"type,omitempty"`
	Path    string       `json:"path,omitempty"`
	Options *DiskOptions `json:"options,omitempty"`
}

// DiskOptions disk options
type DiskOptions struct {
	SizeGB int  `json:"sizeGb,omitempty"`
	Sparse bool `json:"sparse,omitempty"`
}

// +kubebuilder:object:root=true

// Vino is the Schema for the vinoes API
type Vino struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VinoSpec   `json:"spec,omitempty"`
	Status VinoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VinoList contains a list of Vino
type VinoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Vino `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Vino{}, &VinoList{})
}

// VinoStatus defines the observed state of Vino
type VinoStatus struct {
	ConfigMapRef         corev1.ObjectReference `json:"configMapRef,omitempty"`
	Conditions           []Condition            `json:"conditions,omitempty"`
	ConfigMapReady       bool
	VirtualMachinesReady bool
	NetworkingReady      bool
	DaemonSetReady       bool
}

// Condition indicates operational status of VINO CR
type Condition struct {
	Status  corev1.ConditionStatus
	Type    ConditionType
	Reason  string
	Message string
}

// ConditionType type of the condition
type ConditionType string

const (
	ConditionTypeError ConditionType = "Error"
	ConditionTypeInfo  ConditionType = "Info"
	ConditionTypeReady ConditionType = "Ready"
)
