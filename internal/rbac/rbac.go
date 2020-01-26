package rbac

import (
	"fmt"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

//go:generate mockgen -source=rbac.go -destination mocks/enumerator.go -package=rbacmocks

type (
	// Enumerator defines the interface that allows retrieving RBAC roles given
	// zero or more filters
	Enumerator interface {
		EnumberateByRoleBindings(namespace string, filters ...RoleBindingFilter) ([]v1.RoleBinding, error)
	}
	// enumerator is the concrete implementation of the Enumerator interface
	enumerator struct {
		client rbacV1Interface
	}
	// rbacV1Interface is a simplified rbacv1.RbacV1Interface
	rbacV1Interface interface {
		Roles(namespace string) rbacv1.RoleInterface
		RoleBindings(namespace string) rbacv1.RoleBindingInterface
	}
)

// New given a rbacV1Interface returns an Enumerator, or error
func New(client rbacV1Interface) (Enumerator, error) {
	return &enumerator{
		client: client,
	}, nil
}

// EnumberateByRoleBindings returns role bindings that match the given filters
func (e *enumerator) EnumberateByRoleBindings(namespace string, filters ...RoleBindingFilter) ([]v1.RoleBinding, error) {
	roleBindingsOptions := metav1.ListOptions{}
	roleBindings, err := e.client.RoleBindings(namespace).List(roleBindingsOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	filteredRoleBindings := []v1.RoleBinding{}
	for _, roleBinding := range roleBindings.Items {
		for _, filter := range filters {
			if filter(roleBinding) {
				filteredRoleBindings = append(filteredRoleBindings, roleBinding)
				continue
			}
		}
	}

	return filteredRoleBindings, nil
}
