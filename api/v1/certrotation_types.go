package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CertRotationStrategy string

const (
	CertRotationImmediate CertRotationStrategy = "Immediate"
	CertRotationWindowed  CertRotationStrategy = "Windowed"
)

// Weekday is a short weekday name.
// +kubebuilder:validation:Enum=Mon;Tue;Wed;Thu;Fri;Sat;Sun
type Weekday string

// CertRotationWindow defines a maintenance window.
type CertRotationWindow struct {
	// Days is a list like ["Mon","Tue","Wed","Thu","Fri"].
	Days []Weekday `json:"days,omitempty"`

	// +kubebuilder:validation:Pattern=`^([01][0-9]|2[0-3]):[0-5][0-9]$`
	Start string `json:"start"`

	// +kubebuilder:validation:Pattern=`^([01][0-9]|2[0-3]):[0-5][0-9]$`
	End string `json:"end"`
}

// CertRotationPolicy controls restart timing for cert-driven rollouts.
type CertRotationPolicy struct {
	// Strategy controls restart timing.
	// +kubebuilder:validation:Enum=Immediate;Windowed
	// +kubebuilder:default:=Immediate
	Strategy CertRotationStrategy `json:"strategy,omitempty"`

	// UrgentBefore: restart immediately if the *loaded* cert expires within this duration.
	// Format is a Go duration string (e.g. "24h", "15m", "2h30m").
	UrgentBefore *metav1.Duration `json:"urgentBefore,omitempty"`

	// Timezone is an IANA timezone name used to evaluate maintenance windows
	// (e.g. "Europe/London", "America/New_York").
	Timezone string `json:"timezone,omitempty"`

	// Windows defines maintenance windows used when Strategy=Windowed.
	Windows []CertRotationWindow `json:"windows,omitempty"`
}
