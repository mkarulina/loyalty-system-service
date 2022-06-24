package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	mock_authentication "github.com/mkarulina/loyalty-system-service.git/internal/authentication/mocks"
	mock_storage "github.com/mkarulina/loyalty-system-service.git/internal/storage/mocks"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handler_LoginHandler_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	auth.EXPECT().CheckUserData(gomock.Any()).Return(nil)

	reqBody, _ := json.Marshal(loginReq{
		Login:    "testLogin",
		Password: "testPassword",
	})

	r := bytes.NewReader(reqBody)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/user/login", r)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.LoginHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	err = result.Body.Close()
	require.NoError(t, err)
}

func Test_handler_LoginHandler_authError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	auth.EXPECT().CheckUserData(gomock.Any()).Return(errors.New("some error"))

	reqBody, _ := json.Marshal(loginReq{
		Login:    "testLogin",
		Password: "testPassword",
	})

	r := bytes.NewReader(reqBody)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/user/login", r)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.LoginHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusBadRequest, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	err = result.Body.Close()
	require.NoError(t, err)
}
