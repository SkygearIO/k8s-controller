package condition

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ToStatus(b bool) metav1.ConditionStatus {
	if b {
		return metav1.ConditionTrue
	} else {
		return metav1.ConditionFalse
	}
}
