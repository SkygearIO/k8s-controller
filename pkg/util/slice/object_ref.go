package slice

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ContainsObjectReference(slice []corev1.ObjectReference, m metav1.Object) bool {
	for _, elem := range slice {
		if elem.UID == m.GetUID() {
			return true
		}
	}
	return false
}

func RemoveObjectReference(slice []corev1.ObjectReference, m metav1.Object) []corev1.ObjectReference {
	newSlice := []corev1.ObjectReference{}
	for _, elem := range slice {
		if elem.UID != m.GetUID() {
			newSlice = append(newSlice, elem)
		}
	}
	return newSlice
}
