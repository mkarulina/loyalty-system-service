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

func Test_handler_GetOrderHandler_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testTime := time.Now()

	stgResp := []storage.Order{
		{
			UserID:     "1q2w3e4r",
			Number:     "12345",
			Status:     "TESTSTATUS",
			Accrual:    100,
			Withdrawn:  50,
			UploadedAt: testTime,
		}, {
			UserID:     "5t6y7u8i",
			Number:     "67890",
			Status:     "TESTSTATUS",
			Accrual:    200,
			Withdrawn:  100,
			UploadedAt: testTime,
		},
	}

	wantResp, _ := json.Marshal([]orderResp{
		{
			Number:     "12345",
			Status:     "TESTSTATUS",
			Accrual:    100,
			UploadedAt: testTime.Format(time.RFC3339),
		},
		{
			Number:     "67890",
			Status:     "TESTSTATUS",
			Accrual:    200,
			UploadedAt: testTime.Format(time.RFC3339),
		},
	})

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	orderStg.EXPECT().GetUserOrders(gomock.Any()).Return(stgResp, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.GetOrderHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Equal(t, wantResp, body)

	err = result.Body.Close()
	require.NoError(t, err)
}

func Test_handler_GetOrderHandler_storageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	orderStg.EXPECT().GetUserOrders(gomock.Any()).Return(nil, errors.New("some error"))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.GetOrderHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusInternalServerError, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	err = result.Body.Close()
	require.NoError(t, err)
}

func Test_handler_GetOrderHandler_noContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderStg := mock_storage.NewMockOrderStorage(ctrl)
	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	h := NewHandler(orderStg, historyStg, auth)

	orderStg.EXPECT().GetUserOrders(gomock.Any()).Return([]storage.Order{}, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: "testToken",
	})

	handler := http.HandlerFunc(h.GetOrderHandler)
	handler.ServeHTTP(rec, req)

	result := rec.Result()
	require.Equal(t, http.StatusNoContent, result.StatusCode)

	body, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	err = result.Body.Close()
	require.NoError(t, err)
}
