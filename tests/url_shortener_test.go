package tests

import (
	"net/http"
	"net/url"
	"path"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/random"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	// create a httpexpect instance that will send request to the u (full url)
	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(15),
		}).
		WithBasicAuth("myuser", "mypass").
		Expect().
		Status(http.StatusCreated). // Changed to 201
		JSON().Object().
		ContainsKey("alias")
}

//nolint:funlen
func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name           string
		url            string
		alias          string
		error          string
		expectedStatus int
	}{
		{
			name:           "Valid URL",
			url:            gofakeit.URL(),
			alias:          gofakeit.Word() + gofakeit.Word(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid URL",
			url:            "invalid_url",
			alias:          gofakeit.Word(),
			error:          "field URL is not a valid URL",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty Alias",
			url:            gofakeit.URL(),
			alias:          "",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Duplicate Alias",
			url:            gofakeit.URL(),
			alias:          "duplicate123",
			expectedStatus: http.StatusCreated, // First one succeeds
		},
		{
			name:           "Alias Too Short",
			url:            gofakeit.URL(),
			alias:          "ab",
			error:          "field Alias is not valid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Alias Too Long",
			url:            gofakeit.URL(),
			alias:          "thisaliasistoolong",
			error:          "field Alias is not valid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Alias With Special Chars",
			url:            gofakeit.URL(),
			alias:          "test@alias",
			error:          "field Alias is not valid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			// Save
			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth("myuser", "mypass").
				Expect().Status(tc.expectedStatus).
				JSON().Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().Contains(tc.error)
				return
			}

			alias := tc.alias

			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}

			// Redirect
			testRedirect(t, alias, tc.url)

			// Remove
			reqDel := e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth("myuser", "mypass").Expect().
				Status(http.StatusOK).JSON().Object()

			reqDel.Value("status").String().IsEqual("OK")

			// Redirect again - should fail
			testRedirectNotFound(t, alias)
		})
	}
}

// Test duplicate alias specifically
func TestURLShortener_DuplicateAlias(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	alias := "duplicatetest"

	// Save first time - should succeed
	e.POST("/url").
		WithJSON(save.Request{
			URL:   "https://google.com",
			Alias: alias,
		}).
		WithBasicAuth("myuser", "mypass").
		Expect().
		Status(http.StatusCreated).
		JSON().Object().
		Value("alias").String().IsEqual(alias)

	// Save second time with same alias - should fail
	e.POST("/url").
		WithJSON(save.Request{
			URL:   "https://yahoo.com",
			Alias: alias,
		}).
		WithBasicAuth("myuser", "mypass").
		Expect().
		Status(http.StatusConflict).
		JSON().Object().
		Value("error").String().Contains("already exists")

	// Cleanup
	e.DELETE("/url/"+alias).
		WithBasicAuth("myuser", "mypass").
		Expect().
		Status(http.StatusOK)
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, urlToRedirect, redirectedToURL)
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := api.GetRedirect(u.String())
	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}
