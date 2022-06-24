package handlers

import (
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	mock_authentication "github.com/mkarulina/loyalty-system-service.git/internal/authentication/mocks"
	mock_storage "github.com/mkarulina/loyalty-system-service.git/internal/storage/mocks"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handler_GetBalanceHandler_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	orderStg.EXPECT().GetUserBalanceAndWithdrawn(gomock.Any()).Return(float32(500), float32(300), nil)

	wantResp, err := json.Marshal(&balanceResp{Current: 500, Withdrawn: 300})
	if err != nil {
		log.Println(err)
		return
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/balance", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.GetBalanceHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Equal(t, wantResp, body)

	err = result.Body.Close()
	require.NoError(t, err)
}

func Test_handler_GetBalanceHandler_storageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	orderStg.EXPECT().GetUserBalanceAndWithdrawn(gomock.Any()).Return(float32(0), float32(0), errors.New("some error"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/balance", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.GetBalanceHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusInternalServerError, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	err = result.Body.Close()
	require.NoError(t, err)
}
