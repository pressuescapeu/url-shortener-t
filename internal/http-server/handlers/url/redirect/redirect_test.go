package redirect_test

import (
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/redirect"
	"url-shortener/internal/http-server/handlers/url/redirect/mocks"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLGetter

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://www.google.com/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}
			// here, we create a router because redirect uses chi.URLParam, as well as registering the route
			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))
			// here we create a real running http server on a random port that runs the router
			ts := httptest.NewServer(r)
			defer ts.Close() // we shut it down when the test is done
			// here we make a real GET request to http://localhost:random_port/test_alias
			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias) // api.GetRedirect gives the redirect
			require.NoError(t, err)

			// Check the final URL after redirection.
			assert.Equal(t, tc.url, redirectedToURL)

			// assert - fail but continue checking other tests
			// require - fail and stop
		})
	}
}
