package handlers

import (
	"bytes"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	mock_authentication "github.com/mkarulina/loyalty-system-service.git/internal/authentication/mocks"
	mock_storage "github.com/mkarulina/loyalty-system-service.git/internal/storage/mocks"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handler_SendOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	historyStg := mock_storage.NewMockHistoryStorage(ctrl)
	auth := mock_authentication.NewMockAuth(ctrl)

	tests := []struct {
		name           string
		orderNum       string
		orderStg       func() *mock_storage.MockOrderStorage
		wantStatusCode int
		wantResp       []byte
	}{
		{
			name:     "ok",
			orderNum: "9278923470",
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().AddOrderNumber("9278923470", "testToken").Return(nil)
				return orderStg
			},
			wantStatusCode: http.StatusAccepted,
			wantResp:       []byte{},
		},
		{
			name:     "not valid order number",
			orderNum: "12345",
			orderStg: func() *mock_storage.MockOrderStorage {
				return mock_storage.NewMockOrderStorage(ctrl)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
			wantResp:       []byte("Проверьте формат номера заказа"),
		},
		{
			name:     "order storage error",
			orderNum: "12345678903",
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().AddOrderNumber("12345678903", "testToken").Return(errors.New("some error"))
				return orderStg
			},
			wantStatusCode: http.StatusInternalServerError,
			wantResp:       []byte{},
		},
		{
			name:     "order storage duplicate error",
			orderNum: "12345678903",
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().AddOrderNumber("12345678903", "testToken").Return(errors.New("duplicate"))
				return orderStg
			},
			wantStatusCode: http.StatusOK,
			wantResp:       []byte("the order has already been created by the current user"),
		},
		{
			name:     "order storage unique error",
			orderNum: "12345678903",
			orderStg: func() *mock_storage.MockOrderStorage {
				orderStg := mock_storage.NewMockOrderStorage(ctrl)
				orderStg.EXPECT().AddOrderNumber("12345678903", "testToken").Return(errors.New(pgerrcode.UniqueViolation))
				return orderStg
			},
			wantStatusCode: http.StatusConflict,
			wantResp:       []byte("the order has already been created by the another user"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderStg := tt.orderStg()
			h := NewHandler(orderStg, historyStg, auth)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/user/orders", bytes.NewReader([]byte(tt.orderNum)))
			req.Header.Add(`Content-Type`, `text/plain`)
			req.AddCookie(&http.Cookie{
				Name:  "session_token",
				Value: "testToken",
			})

			handler := http.HandlerFunc(h.SendOrderHandler)
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
