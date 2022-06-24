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

func Test_handler_WithdrawHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	tests := []struct {
		name           string
		reqBody        withdrawReq
		orderStg       func() *mock_storage.MockOrderStorage
		wantStatusCode int
		wantResp       []byte
	}{
		{
			name: "ok",
			reqBody: withdrawReq{
				Order: "12345678903",
				Sum:   100,
			},
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().GetUserBalanceAndWithdrawn("testToken").Return(float32(500), float32(200), nil)
				orderStg.EXPECT().AddOrderNumber("12345678903", "testToken").Return(nil)
				orderStg.EXPECT().WithdrawUserPoints("testToken", "12345678903", float32(100)).Return(nil)
				return orderStg
			},
			wantStatusCode: http.StatusOK,
			wantResp:       []byte{},
		},
		{
			name: "balance < withdraw",
			reqBody: withdrawReq{
				Order: "12345678903",
				Sum:   100,
			},
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().GetUserBalanceAndWithdrawn("testToken").Return(float32(500), float32(500), nil)
				return orderStg
			},
			wantStatusCode: http.StatusPaymentRequired,
			wantResp:       []byte("insufficient funds to write off"),
		},
		{
			name: "order storage GetUserBalanceAndWithdrawn error",
			reqBody: withdrawReq{
				Order: "12345678903",
				Sum:   100,
			},
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().GetUserBalanceAndWithdrawn(gomock.Any()).Return(float32(0), float32(0), errors.New("some error"))
				return orderStg
			},
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       []byte{},
		},
		{
			name: "order storage AddOrderNumber error",
			reqBody: withdrawReq{
				Order: "12345678903",
				Sum:   100,
			},
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().GetUserBalanceAndWithdrawn("testToken").Return(float32(500), float32(200), nil)
				orderStg.EXPECT().AddOrderNumber("12345678903", "testToken").Return(errors.New("some error"))
				return orderStg
			},
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       []byte{},
		},
		{
			name: "order storage WithdrawUserPoints error",
			reqBody: withdrawReq{
				Order: "12345678903",
				Sum:   100,
			},
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().GetUserBalanceAndWithdrawn("testToken").Return(float32(500), float32(200), nil)
				orderStg.EXPECT().AddOrderNumber("12345678903", "testToken").Return(nil)
				orderStg.EXPECT().WithdrawUserPoints("testToken", "12345678903", float32(100)).Return(errors.New("some error"))
				return orderStg
			},
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       []byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderStg := tt.orderStg()
			h := NewHandler(orderStg, historyStg, auth)

			reqBody, _ := json.Marshal(tt.reqBody)

			r := bytes.NewReader(reqBody)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/user/balance/withdraw", r)
			req.AddCookie(&http.Cookie{
				Name:  "session_token",
				Value: "testToken",
			})

			handler := http.HandlerFunc(h.WithdrawHandler)
			handler.ServeHTTP(rec, req)

			result := rec.Result()
			require.Equal(t, tt.wantStatusCode, result.StatusCode)

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.Equal(t, tt.wantResp, body)

			err = result.Body.Close()
			require.NoError(t, err)
		})
	}
}
