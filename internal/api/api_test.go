package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/rbac/v1"
	kfake "k8s.io/client-go/kubernetes/fake"

	"github.com/geoah/go-kube-api/internal/rbac"
	"github.com/geoah/go-kube-api/internal/rbac/fixtures"
	rbacmocks "github.com/geoah/go-kube-api/internal/rbac/mocks"
)

const (
	nsDefault = "default"
)

func TestAPI_RbacEnummerateByBindings(t *testing.T) {
	type fields struct {
		rbac func(t *testing.T) rbac.Enumerator
	}
	type args struct {
		requestBody    string
		requestHeaders http.Header
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		testResp func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "filter by both exact and regexp, req/resp json, faked rbac, success",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					fakeClient := kfake.NewSimpleClientset()
					fakeRbac := fakeClient.RbacV1()
					// items out of order
					_, err := fakeRbac.RoleBindings(nsDefault).Create(&fixtures.RoleBindingRole3Subject3and4)
					require.NoError(t, err, "failed to create sample role binding")
					_, err = fakeRbac.RoleBindings(nsDefault).Create(&fixtures.RoleBindingRole1Subject1)
					require.NoError(t, err, "failed to create sample role binding")
					_, err = fakeRbac.RoleBindings(nsDefault).Create(&fixtures.RoleBindingRole2Subject2)
					require.NoError(t, err, "failed to create sample role binding")
					fakeEnumerator, err := rbac.New(fakeRbac)
					require.NoError(t, err)
					return fakeEnumerator
				},
			},
			args: args{
				requestBody: `{"namespace":"default","subjectNames":["subject1","subject2","subject[3,4]"]}`,
				requestHeaders: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				expResp := []v1.RoleBinding{
					// items should be ordered
					fixtures.RoleBindingRole1Subject1,
					fixtures.RoleBindingRole2Subject2,
					fixtures.RoleBindingRole3Subject3and4,
				}
				resp := []v1.RoleBinding{}
				respBody, _ := ioutil.ReadAll(rr.Body)
				err := json.Unmarshal(respBody, &resp)
				require.NoError(t, err, "could not unmarshal resp")
				assert.Equal(t, expResp, resp)
			},
		},
		{
			name: "filter by both exact and regexp, req/resp json, mocked rbac, success",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					ctrl := gomock.NewController(t)
					mockEnumerator := rbacmocks.NewMockEnumerator(ctrl)
					mockEnumerator.EXPECT().EnumberateByRoleBindings(
						nsDefault,
						gomock.Any(),
					).Return([]v1.RoleBinding{
						// items out of order
						fixtures.RoleBindingRole2Subject2,
						fixtures.RoleBindingRole3Subject3and4,
						fixtures.RoleBindingRole1Subject1,
					}, nil)
					return mockEnumerator
				},
			},
			args: args{
				requestBody: `{"namespace":"default","subjectNames":["subject1","subject2","subject[3,4]"]}`,
				requestHeaders: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				expResp := []v1.RoleBinding{
					// items should be ordered
					fixtures.RoleBindingRole1Subject1,
					fixtures.RoleBindingRole2Subject2,
					fixtures.RoleBindingRole3Subject3and4,
				}
				resp := []v1.RoleBinding{}
				respBody, _ := ioutil.ReadAll(rr.Body)
				err := json.Unmarshal(respBody, &resp)
				require.NoError(t, err, "could not unmarshal resp")
				assert.Equal(t, expResp, resp)
			},
		},
		{
			name: "filter by both exact and regexp, req/resp yaml, mocked rbac, success",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					ctrl := gomock.NewController(t)
					mockEnumerator := rbacmocks.NewMockEnumerator(ctrl)
					mockEnumerator.EXPECT().EnumberateByRoleBindings(
						nsDefault,
						gomock.Any(),
					).Return([]v1.RoleBinding{
						// items out of order
						fixtures.RoleBindingRole3Subject3and4,
						fixtures.RoleBindingRole1Subject1,
						fixtures.RoleBindingRole2Subject2,
					}, nil)
					return mockEnumerator
				},
			},
			args: args{
				requestBody: func() string {
					req := rbacEnumerateByBindingsRequest{
						Namespace: "default",
						SubjectNames: []string{
							"subject1",
							"subject2",
							"subject[3,4]",
						},
					}
					b, _ := yaml.Marshal(req)
					return string(b)
				}(),
				requestHeaders: http.Header{
					"Content-Type": []string{"application/x-yaml"},
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				resp := []v1.RoleBinding{}
				respBody, _ := ioutil.ReadAll(rr.Body)
				err := yaml.Unmarshal(respBody, &resp)
				require.NoError(t, err, "could not unmarshal resp")
				// the unmarshalled yaml nested structures are a bit different
				// for example, instead of the expected
				// `Annotations:map[string]string(nil)`
				// we get
				// `Annotations:map[string]string{}`
				// and it messes with
				// `assert.Equal(t, expResp, resp)`
				// so for now we'll just check their order
				assert.Len(t, resp, 3)
				assert.Equal(t, resp[0].RoleRef.Name, "role1")
				assert.Equal(t, resp[1].RoleRef.Name, "role2")
				assert.Equal(t, resp[2].RoleRef.Name, "role3")
			},
		},
		{
			name: "filter by regexp, rbac error, failure",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					ctrl := gomock.NewController(t)
					mockEnumerator := rbacmocks.NewMockEnumerator(ctrl)
					mockEnumerator.EXPECT().EnumberateByRoleBindings(
						nsDefault,
						gomock.Any(),
					).Return(nil, errors.New("some error"))
					return mockEnumerator
				},
			},
			args: args{
				requestBody: `{"namespace":"default","subjectNames":["subject[3,4]"]}`,
				requestHeaders: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				respBody, _ := ioutil.ReadAll(rr.Body)
				assert.Contains(t, string(respBody), "could not retrieve")
			},
		},
		{
			name: "filter by invalid regexp, failure",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					return nil
				},
			},
			args: args{
				requestBody: `{"namespace":"default","subjectNames":["[["]}`,
				requestHeaders: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				respBody, _ := ioutil.ReadAll(rr.Body)
				assert.Contains(t, string(respBody), "invalid regular expression")
			},
		},
		{
			name: "invalid request body, failure",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					return nil
				},
			},
			args: args{
				requestBody: `{"a`,
				requestHeaders: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				respBody, _ := ioutil.ReadAll(rr.Body)
				assert.Contains(t, string(respBody), "could not parse")
			},
		},
		{
			name: "missing namespace, failure",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					return nil
				},
			},
			args: args{
				requestBody: `{"namespace":"","subjectNames":["subject1"]}`,
				requestHeaders: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				respBody, _ := ioutil.ReadAll(rr.Body)
				assert.Contains(t, string(respBody), "missing namespace")
			},
		},
		{
			name: "missing subject names, failure",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					return nil
				},
			},
			args: args{
				requestBody: `{"namespace":"default","subjectNames":[]}`,
				requestHeaders: http.Header{
					"Content-Type": []string{"application/json"},
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				respBody, _ := ioutil.ReadAll(rr.Body)
				assert.Contains(t, string(respBody), "missing subject names")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rbacMock := tt.fields.rbac(t)
			api, err := New(rbacMock)
			require.NoError(t, err, "failed to create new api")

			r := gin.Default()
			r.POST("/", api.RbacEnummerateByBindings)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/", strings.NewReader(tt.args.requestBody))
			req.Header = tt.args.requestHeaders
			r.ServeHTTP(w, req)
			tt.testResp(t, w)
		})
	}
}

func TestAPI_Health(t *testing.T) {
	type fields struct {
		rbac func(t *testing.T) rbac.Enumerator
	}
	tests := []struct {
		name     string
		fields   fields
		testResp func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					ctrl := gomock.NewController(t)
					mockEnumerator := rbacmocks.NewMockEnumerator(ctrl)
					mockEnumerator.EXPECT().EnumberateByRoleBindings(
						"",
						gomock.Any(),
					).Return([]v1.RoleBinding{}, nil)
					return mockEnumerator
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "failure",
			fields: fields{
				rbac: func(t *testing.T) rbac.Enumerator {
					ctrl := gomock.NewController(t)
					mockEnumerator := rbacmocks.NewMockEnumerator(ctrl)
					mockEnumerator.EXPECT().EnumberateByRoleBindings(
						"",
						gomock.Any(),
					).Return(nil, errors.New("some error"))
					return mockEnumerator
				},
			},
			testResp: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rbacMock := tt.fields.rbac(t)
			api, err := New(rbacMock)
			require.NoError(t, err, "failed to create new api")

			r := gin.Default()
			r.GET("/", api.Health)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			r.ServeHTTP(w, req)
			tt.testResp(t, w)
		})
	}
}
