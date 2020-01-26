package rbac

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/rbac/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	kfake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/geoah/go-kube-api/internal/rbac/fixtures"
)

const (
	nsDefault = "default"
)

var (
	rolesResource = schema.GroupVersionResource{
		Group:    "rbac.authorization.k8s.io",
		Version:  "v1",
		Resource: "roles",
	}
	rolesKind = schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "Role",
	}
)

func Test_enumerator_EnumberateByRoleBindings(t *testing.T) {
	type fields struct {
		client rbacV1Interface
	}
	type args struct {
		namespace string
		filters   []RoleBindingFilter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []v1.RoleBinding
		wantErr bool
	}{
		{
			name: "filter by both exact and regexp, success",
			fields: fields{
				client: func() rbacV1Interface {
					fakeClient := kfake.NewSimpleClientset()
					fakeRbac := fakeClient.RbacV1()
					_, err := fakeRbac.RoleBindings(nsDefault).Create(&fixtures.RoleBindingRole1Subject1)
					require.NoError(t, err, "failed to create sample role binding")
					_, err = fakeRbac.RoleBindings(nsDefault).Create(&fixtures.RoleBindingRole2Subject2)
					require.NoError(t, err, "failed to create sample role binding")
					_, err = fakeRbac.RoleBindings(nsDefault).Create(&fixtures.RoleBindingRole3Subject3and4)
					require.NoError(t, err, "failed to create sample role binding")
					return fakeRbac
				}(),
			},
			args: args{
				namespace: nsDefault,
				filters: []RoleBindingFilter{
					FilterBySubjectName("subject1"),
					FilterBySubjectName("subject2"),
					FilterBySubjectNameRegex(*regexp.MustCompile("subject[3,4]")),
				},
			},
			want: []v1.RoleBinding{
				fixtures.RoleBindingRole1Subject1,
				fixtures.RoleBindingRole2Subject2,
				fixtures.RoleBindingRole3Subject3and4,
			},
			wantErr: false,
		},
		{
			name: "client error, fails",
			fields: fields{
				client: func() rbacV1Interface {
					fakeClient := kfake.NewSimpleClientset()
					fakeClient.ReactionChain = []ktesting.Reactor{}
					fakeClient.AddReactor("*", "*", func(action ktesting.Action) (bool, kruntime.Object, error) {
						return true, nil, errors.New("something went wrong")
					})
					fakeRbac := fakeClient.RbacV1()
					return fakeRbac
				}(),
			},
			args: args{
				namespace: nsDefault,
				filters: []RoleBindingFilter{
					FilterBySubjectName("does-not-matter"),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := New(tt.fields.client)
			require.NoError(t, err, "failed to create new rbac enumerator")
			got, err := e.EnumberateByRoleBindings(tt.args.namespace, tt.args.filters...)
			if tt.wantErr {
				require.Error(t, err, "expected error but got none")
			} else {
				require.NoError(t, err, "did not expect error")
			}
			assert.Equal(t, tt.want, got, "response did not match expectation")
		})
	}
}
