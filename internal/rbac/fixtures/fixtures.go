package fixtures

import (
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nsDefault = "default"
)

var (
	// RoleBindingRole1Subject1 sample role binding
	RoleBindingRole1Subject1 = v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "role1-for-subject1",
			Namespace: nsDefault,
		},
		Subjects: []v1.Subject{
			v1.Subject{
				Kind: "User",
				Name: "subject1",
			},
		},
		RoleRef: v1.RoleRef{
			Name: "role1",
		},
	}
	// RoleBindingRole2Subject2 sample role binding
	RoleBindingRole2Subject2 = v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "role2-for-subject2",
			Namespace: nsDefault,
		},
		Subjects: []v1.Subject{
			v1.Subject{
				Kind: "User",
				Name: "subject2",
			},
		},
		RoleRef: v1.RoleRef{
			Name: "role2",
		},
	}
	// RoleBindingRole3Subject3and4 sample role binding
	RoleBindingRole3Subject3and4 = v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "role3-for-subject3and4",
			Namespace: nsDefault,
		},
		Subjects: []v1.Subject{
			v1.Subject{
				Kind: "User",
				Name: "subject3",
			},
			v1.Subject{
				Namespace: nsDefault,
				Name:      "subject4",
			},
		},
		RoleRef: v1.RoleRef{
			Name: "role3",
		},
	}
)
