package redirect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/redirect"
	"url-shortener/internal/http-server/handlers/url/redirect/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

func TestRedirectHandler(t *testing.T) { // Fixed name!
	cases := []struct {
		name           string
		alias          string
		url            string
		mockError      error
		expectedStatus int
		expectedURL    string // For checking Location header
	}{
		{
			name:           "Success",
			alias:          "test_alias",
			url:            "https://google.com",
			mockError:      nil,
			expectedStatus: http.StatusFound, // 302
			expectedURL:    "https://google.com",
		},
		{
			name:           "Empty alias",
			alias:          "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest, // 400
		},
		{
			name:           "URL not found",
			alias:          "notfound",
			mockError:      storage.ErrURLNotFound,
			expectedStatus: http.StatusNotFound, // 404
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlGetterMock := mocks.NewURLGetter(t)

			// Only set up mock if alias is not empty
			if tc.alias != "" {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).
					Once()
			}

			// Create chi context with alias
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)

			// Create request
			req, err := http.NewRequest(http.MethodGet, "/"+tc.alias, nil)
			require.NoError(t, err)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create handler and recorder
			handler := redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock)
			rr := httptest.NewRecorder()

			// Execute
			handler.ServeHTTP(rr, req)

			// Check status code
			require.Equal(t, tc.expectedStatus, rr.Code)

			// For successful redirect, check Location header
			if tc.expectedStatus == http.StatusFound {
				location := rr.Header().Get("Location")
				require.Equal(t, tc.expectedURL, location)
			}
		})
	}
}
