package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/save/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

// testing - we know if function actually works or not bc we test it not manually but with code (duh)
// unit testing - testing individual pieces (functions) in isolation
// integration testing - testing how all those functions work together
// ok so when we test save.New() we don't actually connect to db, insert data, all that
// so we use mock - and we generate that mock by using this:
//go:generate go run github.com/vektra/mockery/v2@latest --name=URLSaver

func TestSaveHandler(t *testing.T) {
	// this is a table-driven set, meaning that instead of writing 5 separate functions,
	// we define the test case with 5 attributes
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		// this is a successful test case - user sends a valid alias with a valid url and it's all good
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
		},
		// this is an empty alias test case (duh)
		// we generate a random alias for this so the code still works
		{
			name:  "Empty alias",
			alias: "",
			url:   "https://google.com",
		},
		// user sends alias but no url - so we expect the message of respError
		{
			name:      "Empty URL",
			url:       "",
			alias:     "some_alias",
			respError: "field URL is a required field",
		},
		// this is a check for the validation of the URL
		{
			name:      "Invalid URL",
			url:       "some invalid URL",
			alias:     "some_alias",
			respError: "field URL is not a valid URL",
		},
		// everything is valid, but in case there is some unexpected error
		{
			name:      "SaveURL Error",
			alias:     "test_alias",
			url:       "https://google.com",
			respError: "failed to add url",
			mockError: errors.New("unexpected error"),
		},
	}

	// ok so here we go through the test cases
	for _, tc := range cases {
		tc := tc // here we copy the test case for parallel tests

		// t.Run() creates a sub-test with the name in output and runs it in parallel
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // parallel is here btw
			// here we create a fake object, in this case, fake db object
			urlSaverMock := mocks.NewURLSaver(t)
			// here we program the mock so that
			// we set up the mock only if we expect success or if we're testing a database error
			// success - because respError is empty
			// database error - mockError is not nil
			// if the validation fails and url is empty, then we don't even call database
			if tc.respError == "" || tc.mockError != nil {
				// this line is - when SaveURL is called with the url from the test case,
				// and some string - might be empty alias btw
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					// we return id = 1 and the error in the test case
					Return(int64(1), tc.mockError).
					// the call should be only once, if not, the test case is failed
					Once()
			}
			// we create the save handler, pass it to the logger that discards logs, and pass the mock
			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)
			// here we create a fake http request
			// we build the json string
			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)
			// we create a fake post request with the json
			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err) // in case of creating request failed, we stop the test
			// here, we create a fake response recorder
			rr := httptest.NewRecorder() // this is a fake http.ResponseWriter basically
			handler.ServeHTTP(rr, req)   // this runs our handler with the fake request
			// now here, rr contains the response - status code, body, headers
			require.Equal(t, rr.Code, http.StatusOK) // here we check if the status code is 200
			// get the response body as string
			body := rr.Body.String()
			// save the json as response struct
			var resp save.Response
			// check on  the error message matching what we expect
			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

			// TODO: add more checks
		})
	}
}
