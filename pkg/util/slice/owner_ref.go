package slice

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ContainsOwnerReference(slice []metav1.OwnerReference, m metav1.Object) bool {
	for _, elem := range slice {
		if elem.UID == m.GetUID() {
			return true
		}
	}
	return false
}

func RemoveOwnerReference(slice []metav1.OwnerReference, m metav1.Object) []metav1.OwnerReference {
	newSlice := []metav1.OwnerReference{}
	for _, elem := range slice {
		if elem.UID != m.GetUID() {
			newSlice = append(newSlice, elem)
		}
	}
	return newSlice
}
