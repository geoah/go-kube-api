package api

import (
	"net/http"
	"regexp"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/geoah/go-kube-api/internal/rbac"
)

var (
	namespaceRegexp                   = regexp.MustCompile("^[0-9A-Za-z]+$")
	roleBindingExactSubjectNameRegexp = regexp.MustCompile("^[0-9A-Za-z]+$")
)

type (
	// API provides the handlers for the echo HTTP server
	API struct {
		rbac rbac.Enumerator
	}
	// rbacEnumerateByBindingsRequest
	rbacEnumerateByBindingsRequest struct {
		Namespace    string   `json:"namespace" yaml:"namespace"`
		SubjectNames []string `json:"subjectNames" yaml:"subjectNames"`
	}
)

// New API given an RbacEnumerator
func New(rbac rbac.Enumerator) (*API, error) {
	return &API{
		rbac: rbac,
	}, nil
}

// RbacEnummerateByBindings handles requests to enumerate role bindings filtered
// by subject names
func (api API) RbacEnummerateByBindings(c *gin.Context) {
	// construct request
	req := rbacEnumerateByBindingsRequest{}
	if err := c.Bind(&req); err != nil {
		c.Render(http.StatusBadRequest, renderer(c, "could not parse request"))
		return
	}

	// validate namespace
	if req.Namespace == "" {
		c.Render(http.StatusBadRequest, renderer(c, "missing namespace in request"))
		return
	}

	// validate subject names
	if len(req.SubjectNames) == 0 {
		c.Render(http.StatusBadRequest, renderer(c, "missing subject names in request"))
		return
	}

	// go through subject names and construct rbac filters
	filters := make([]rbac.RoleBindingFilter, len(req.SubjectNames))
	for i, subjectName := range req.SubjectNames {
		// check if subject name is simple enough to be an exact match
		if roleBindingExactSubjectNameRegexp.MatchString(subjectName) {
			filters[i] = rbac.FilterBySubjectName(subjectName)
			continue
		}
		// else we assume it's a regular expression which needs to be compiled
		subjectNameRegexp, err := regexp.Compile(subjectName)
		if err != nil {
			c.Render(http.StatusBadRequest, renderer(c, "invalid regular expression or subject name"))
			return
		}
		filters[i] = rbac.FilterBySubjectNameRegex(*subjectNameRegexp)
	}

	// retrieve filtered role bindings
	roleBindings, err := api.rbac.EnumberateByRoleBindings(req.Namespace, filters...)
	if err != nil {
		c.Render(http.StatusInternalServerError, renderer(c, "could not retrieve role bindings"))
		return
	}

	// sort role bindings by role name
	sort.Slice(roleBindings, func(i, j int) bool {
		return roleBindings[i].RoleRef.Name < roleBindings[j].RoleRef.Name
	})

	// return response
	c.Render(http.StatusOK, renderer(c, roleBindings))
}
