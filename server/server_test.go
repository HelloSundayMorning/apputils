package server

import (
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppServer_AddAuthorizedRoute(t *testing.T) {

	srv := NewServer("APPID", 8000)

	err := srv.AddAuthorizedRoute("/", "GET", []string{"ROLE1", "ROLE2"}, func(writer http.ResponseWriter, request *http.Request) {

	})

	assert.Nil(t, err)

	err = srv.AddAuthorizedRoute("/NOROLE", "GET", []string{}, func(writer http.ResponseWriter, request *http.Request) {

	})

	assert.Nil(t, err)

	type testDef struct {
		Req *http.Request
		ExpectedStatus int
	}

	reqFunc := func(method, path, userID string, roles string) *http.Request {
		r := httptest.NewRequest(method,path, nil)
		r.Header.Add(appctx.AuthorizedUserRolesHeader, roles)
		r.Header.Add(appctx.AuthorizedUserIDHeader, userID)

		return r
	}

	Tests := []testDef{
		{reqFunc("GET","/APPID/", "", ""), http.StatusUnauthorized},
		{reqFunc("GET","/APPID/", "", "ROLE1"), http.StatusUnauthorized},
		{reqFunc("GET","/APPID/", "UserID", ""), http.StatusForbidden},
		{reqFunc("GET","/APPID/", "UserID", "ROLE1"), http.StatusOK},
		{reqFunc("GET","/APPID/", "UserID", "ROLE2"), http.StatusOK},
		{reqFunc("GET","/APPID/", "UserID", "ROLE3"), http.StatusForbidden},
		{reqFunc("GET","/APPID/", "UserID", "ROLE1,ROLE3"), http.StatusOK},
		{reqFunc("GET","/APPID/", "UserID", "ROLE3,ROLE1"), http.StatusOK},
		{reqFunc("GET","/APPID/", "UserID", "ROLE1,ROLE2"), http.StatusOK},
		{reqFunc("GET","/APPID/", "UserID", "ROLE3,ROLE4"), http.StatusForbidden},
		{reqFunc("GET","/APPID/NOROLE", "", ""), http.StatusUnauthorized},
		{reqFunc("GET","/APPID/NOROLE", "", "ROLE1"), http.StatusUnauthorized},
		{reqFunc("GET","/APPID/NOROLE", "UserID", ""), http.StatusOK},
	}

	for idx, test := range Tests{

		wri := httptest.NewRecorder()

		srv.Handler.ServeHTTP(wri, test.Req)

		resp := wri.Result()
		assert.Equalf(t, test.ExpectedStatus, resp.StatusCode, "Failed test %d", idx)
	}
}


