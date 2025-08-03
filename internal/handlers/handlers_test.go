package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/etoneja/go-gophermart/internal/models"
	"github.com/etoneja/go-gophermart/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBalanceHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockServicer(ctrl)
	hs := NewHandlers(mockSvc, zerolog.Nop())

	testUser := &models.UserModel{UUID: "fakeUUID"}
	balance := &models.BalanceModel{
		Current:   105,
		Withdrawn: 50,
	}

	mockSvc.EXPECT().
		GetUserBalance(gomock.Any(), testUser.UUID).
		Return(balance, nil).
		Times(1)

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user", testUser)
	c.Request = req

	hs.GetBalanceHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	expectedResponse := balance.ToResponse()

	var response models.BalanceResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.Current, response.Current)
	assert.Equal(t, expectedResponse.Withdrawn, response.Withdrawn)
}
