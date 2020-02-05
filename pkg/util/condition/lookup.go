package condition

import "github.com/skygeario/k8s-controller/api"

func Lookup(conds []api.Condition, condType string) *api.Condition {
	for _, cond := range conds {
		if cond.Type == condType {
			return &cond
		}
	}
	return nil
}
