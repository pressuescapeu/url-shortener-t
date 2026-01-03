package delete_test

import (
	"context"
	"encoding/json"
	"errors"

	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLDeleter
func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name           string
		alias          string
		respError      string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Success",
			alias:          "test_alias",
			respError:      "",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty alias",
			alias:          "",
			respError:      "not found",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "URL not found",
			alias:          "no_url",
			respError:      "url not found",
			mockError:      storage.ErrNoURLDeleted,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "DeleteURL error",
			alias:          "some_alias",
			respError:      "failed to delete url",
			mockError:      errors.New("unexpected error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Alias too short",
			alias:          "ab",
			respError:      "invalid alias",
			mockError:      nil, // Mock not called
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Alias too long",
			alias:          "this_is_extremely_long_alias_exceeding_reasonable_limits_for_url_shortener",
			respError:      "invalid alias",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			urlDeleterMock := mocks.NewURLDeleter(t)
			if tc.alias != "" && len(tc.alias) >= 3 && len(tc.alias) <= 15 { // empty alias case does not call DeleteURL
				urlDeleterMock.On("DeleteURL", tc.alias).Return(tc.mockError).Once()
			}
			// create chi's route context that hold the url params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias) // add alias param
			// create a basic DELETE request
			req, err := http.NewRequest(http.MethodDelete, "/url/"+tc.alias, nil)
			require.NoError(t, err)
			// attach the chi context to the request
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			// create delete handler
			handler := delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock)
			// run the handler
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// check the status code
			require.Equal(t, tc.expectedStatus, rr.Code)
			// check response body
			body := rr.Body.String()
			var resp delete.Response
			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
