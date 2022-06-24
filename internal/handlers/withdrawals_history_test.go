package handlers

import (
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	mock_authentication "github.com/mkarulina/loyalty-system-service.git/internal/authentication/mocks"
	"github.com/mkarulina/loyalty-system-service.git/internal/storage"
	mock_storage "github.com/mkarulina/loyalty-system-service.git/internal/storage/mocks"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_handler_GetWithdrawalsHistoryHandler_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTime := time.Now()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	historyStg.EXPECT().GetWithdrawalsHistory(gomock.Any()).Return([]storage.Withdrawn{
		{
			UserID:      "1q2w3e4r",
			OrderNumber: "12345",
			Sum:         111,
			ProcessedAt: testTime,
		},
		{
			UserID:      "5t6y7u8i",
			OrderNumber: "67890",
			Sum:         222,
			ProcessedAt: testTime,
		},
	}, nil)

	wantResp, _ := json.Marshal([]withdrawalsHistoryResp{
		{
			Order:       "12345",
			Sum:         111,
			ProcessedAt: testTime,
		},
		{
			Order:       "67890",
			Sum:         222,
			ProcessedAt: testTime,
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/balance/withdrawals", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.GetWithdrawalsHistoryHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Equal(t, wantResp, body)

	err = result.Body.Close()
	require.NoError(t, err)
}

func Test_handler_GetWithdrawalsHistoryHandler_storageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	historyStg.EXPECT().GetWithdrawalsHistory(gomock.Any()).Return(nil, errors.New("some error"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/balance/withdrawals", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.GetWithdrawalsHistoryHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusInternalServerError, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	err = result.Body.Close()
	require.NoError(t, err)
}

func Test_handler_GetWithdrawalsHistoryHandler_noContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	historyStg.EXPECT().GetWithdrawalsHistory(gomock.Any()).Return([]storage.Withdrawn{}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/balance/withdrawals", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.GetWithdrawalsHistoryHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusNoContent, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	err = result.Body.Close()
	require.NoError(t, err)
}
