package rbac

import (
	"regexp"

	v1 "k8s.io/api/rbac/v1"
)

type (
	// RoleBindingFilter for the RBAC enumerator
	RoleBindingFilter func(role v1.RoleBinding) bool
)

// FilterBySubjectName allows filtering rolebindings by their exact subject name
func FilterBySubjectName(subjectName string) RoleBindingFilter {
	return func(roleBinding v1.RoleBinding) bool {
		for _, subject := range roleBinding.Subjects {
			if subject.Name == subjectName {
				return true
			}
		}
		return false
	}
}

// FilterBySubjectNameRegex allows filtering rolebindings by a regular expression
func FilterBySubjectNameRegex(subjectNameRegexp regexp.Regexp) RoleBindingFilter {
	return func(roleBinding v1.RoleBinding) bool {
		for _, subject := range roleBinding.Subjects {
			if subjectNameRegexp.MatchString(subject.Name) {
				return true
			}
		}
		return false
	}
}
