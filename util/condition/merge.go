package condition

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/skygeario/k8s-controller/api"
)

func MergeFrom(newConds, oldConds []api.Condition) {
	for i, cond := range newConds {
		updated := false
		for _, old := range oldConds {
			if old.Type != cond.Type {
				continue
			}

			if cond.Status != old.Status {
				cond.LastTransitionTime = metav1.Now()
			} else {
				cond.LastTransitionTime = old.LastTransitionTime
				if cond.Message == "" {
					cond.Message = old.Message
					cond.Reason = old.Reason
				}
			}
			updated = true
			break
		}
		if !updated {
			cond.LastTransitionTime = metav1.Now()
		}
		newConds[i] = cond
	}
}
