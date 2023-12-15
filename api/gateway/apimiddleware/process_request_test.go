package apimiddleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/api/grpc"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

type testRequestContainer struct {
	TestString    string
	TestHexString string `hex:"true"`
}

func defaultRequestContainer() *testRequestContainer {
	return &testRequestContainer{
		TestString:    "test string",
		TestHexString: "0x666F6F", // hex encoding of "foo"
	}
}

type testResponseContainer struct {
	TestString string
	TestHex    string `hex:"true"`
	TestEnum   string `enum:"true"`
	TestTime   string `time:"true"`
}

func defaultResponseContainer() *testResponseContainer {
	return &testResponseContainer{
		TestString: "test string",
		TestHex:    "Zm9v", // base64 encoding of "foo"
		TestEnum:   "Test Enum",
		TestTime:   "2006-01-02T15:04:05Z",
	}
}

type testErrorJSON struct {
	Message     string
	Code        int
	CustomField string
}

// StatusCode returns the error's underlying error code.
func (e *testErrorJSON) StatusCode() int {
	return e.Code
}

// Msg returns the error's underlying message.
func (e *testErrorJSON) Msg() string {
	return e.Message
}

// SetCode sets the error's underlying error code.
func (e *testErrorJSON) SetCode(code int) {
	e.Code = code
}

// SetMsg sets the error's underlying message.
func (e *testErrorJSON) SetMsg(msg string) {
	e.Message = msg
}

func TestDeserializeRequestBodyIntoContainer(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var bodyJSON bytes.Buffer
		err := json.NewEncoder(&bodyJSON).Encode(defaultRequestContainer())
		require.NoError(t, err)

		container := &testRequestContainer{}
		errJson := DeserializeRequestBodyIntoContainer(&bodyJSON, container)
		require.Equal(t, true, errJson == nil)
		assert.Equal(t, "test string", container.TestString)
	})

	t.Run("error", func(t *testing.T) {
		var bodyJSON bytes.Buffer
		bodyJSON.Write([]byte("foo"))
		errJson := DeserializeRequestBodyIntoContainer(&bodyJSON, &testRequestContainer{})
		require.NotNil(t, errJson)
		assert.Equal(t, true, strings.Contains(errJson.Msg(), "could not decode request body"))
		assert.Equal(t, http.StatusInternalServerError, errJson.StatusCode())
	})

	t.Run("unknown field", func(t *testing.T) {
		var bodyJSON bytes.Buffer
		bodyJSON.Write([]byte("{\"foo\":\"foo\"}"))
		errJSON := DeserializeRequestBodyIntoContainer(&bodyJSON, &testRequestContainer{})
		require.NotNil(t, errJSON)
		assert.Equal(t, true, strings.Contains(errJSON.Msg(), "could not decode request body"))
		assert.Equal(t, http.StatusBadRequest, errJSON.StatusCode())
	})
}

func TestProcessRequestContainerFields(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		container := defaultRequestContainer()

		errJSON := ProcessRequestContainerFields(container)
		require.Equal(t, true, errJSON == nil)
		assert.Equal(t, "Zm9v", container.TestHexString)
	})

	t.Run("error", func(t *testing.T) {
		errJSON := ProcessRequestContainerFields("foo")
		require.NotNil(t, errJSON)
		assert.Equal(t, true, strings.Contains(errJSON.Msg(), "could not process request data"))
		assert.Equal(t, http.StatusInternalServerError, errJSON.StatusCode())
	})
}

func TestSetRequestBodyToRequestContainer(t *testing.T) {
	var body bytes.Buffer
	request := httptest.NewRequest("GET", "http://foo.example", &body)

	errJSON := SetRequestBodyToRequestContainer(defaultRequestContainer(), request)
	require.Equal(t, true, errJSON == nil)
	container := &testRequestContainer{}
	require.NoError(t, json.NewDecoder(request.Body).Decode(container))
	assert.Equal(t, "test string", container.TestString)
	contentLengthHeader, ok := request.Header["Content-Length"]
	require.Equal(t, true, ok)
	require.Equal(t, 1, len(contentLengthHeader), "wrong number of header values")
	assert.Equal(t, "55", contentLengthHeader[0])
	assert.Equal(t, int64(55), request.ContentLength)
}

func TestPrepareRequestForProxying(t *testing.T) {
	middleware := &APIProxyMiddleware{
		GatewayAddress: "http://gateway.example",
	}
	// We will set some params to make the request more interesting.
	endpoint := Endpoint{
		Path:               "/{url_param}",
		RequestURLLiterals: []string{"url_param"},
		RequestQueryParams: []QueryParam{{Name: "query_param"}},
	}
	var body bytes.Buffer
	request := httptest.NewRequest("GET", "http://foo.example?query_param=bar", &body)

	errJSON := middleware.PrepareRequestForProxying(endpoint, request)
	require.Equal(t, true, errJSON == nil)
	assert.Equal(t, "http", request.URL.Scheme)
	assert.Equal(t, middleware.GatewayAddress, request.URL.Host)
	assert.Equal(t, "", request.RequestURI)
}

func TestReadGrpcResponseBody(t *testing.T) {
	var b bytes.Buffer
	b.Write([]byte("foo"))

	body, jsonERR := ReadGrpcResponseBody(&b)
	require.Equal(t, true, jsonERR == nil)
	assert.Equal(t, "foo", string(body))
}

func TestHandleGrpcResponseError(t *testing.T) {
	response := &http.Response{
		StatusCode: 400,
		Header: http.Header{
			"Foo": []string{"foo"},
			"Bar": []string{"bar"},
		},
	}
	writer := httptest.NewRecorder()
	errJSON := &testErrorJSON{
		Message: "foo",
		Code:    400,
	}
	b, err := json.Marshal(errJSON)
	require.NoError(t, err)

	hasError, e := HandleGrpcResponseError(errJSON, response, b, writer)
	require.Equal(t, true, e == nil)
	assert.Equal(t, true, hasError)
	v, ok := writer.Header()["Foo"]
	require.Equal(t, true, ok, "header not found")
	require.Equal(t, 1, len(v), "wrong number of header values")
	assert.Equal(t, "foo", v[0])
	v, ok = writer.Header()["Bar"]
	require.Equal(t, true, ok, "header not found")
	require.Equal(t, 1, len(v), "wrong number of header values")
	assert.Equal(t, "bar", v[0])
	assert.Equal(t, 400, errJSON.StatusCode())
}

func TestGrpcResponseIsEmpty(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, true, GrpcResponseIsEmpty(nil))
	})
	t.Run("empty_slice", func(t *testing.T) {
		assert.Equal(t, true, GrpcResponseIsEmpty(make([]byte, 0)))
	})
	t.Run("empty_brackets", func(t *testing.T) {
		assert.Equal(t, true, GrpcResponseIsEmpty([]byte("{}")))
	})
	t.Run("non_empty", func(t *testing.T) {
		assert.Equal(t, false, GrpcResponseIsEmpty([]byte("{\"foo\":\"bar\"})")))
	})
}

func TestDeserializeGrpcResponseBodyIntoContainer(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		body, err := json.Marshal(defaultRequestContainer())
		require.NoError(t, err)

		container := &testRequestContainer{}
		errJSON := DeserializeGrpcResponseBodyIntoContainer(body, container)
		require.Equal(t, true, errJSON == nil)
		assert.Equal(t, "test string", container.TestString)
	})

	t.Run("error", func(t *testing.T) {
		var bodyJSON bytes.Buffer
		bodyJSON.Write([]byte("foo"))
		errJson := DeserializeGrpcResponseBodyIntoContainer(bodyJSON.Bytes(), &testRequestContainer{})
		require.NotNil(t, errJson)
		assert.Equal(t, true, strings.Contains(errJson.Msg(), "could not unmarshal response"))
		assert.Equal(t, http.StatusInternalServerError, errJson.StatusCode())
	})
}

func TestProcessMiddlewareResponseFields(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		container := defaultResponseContainer()

		errJSON := ProcessMiddlewareResponseFields(container)
		require.Equal(t, true, errJSON == nil)
		assert.Equal(t, "0x666f6f", container.TestHex)
		assert.Equal(t, "test enum", container.TestEnum)
		assert.Equal(t, "1136214245", container.TestTime)
	})

	t.Run("error", func(t *testing.T) {
		errJSON := ProcessMiddlewareResponseFields("foo")
		require.NotNil(t, errJSON)
		assert.Equal(t, true, strings.Contains(errJSON.Msg(), "could not process response data"))
		assert.Equal(t, http.StatusInternalServerError, errJSON.StatusCode())
	})
}

func TestSerializeMiddlewareResponseIntoJson(t *testing.T) {
	container := defaultResponseContainer()
	j, errJSON := SerializeMiddlewareResponseIntoJSON(container)
	assert.Equal(t, true, errJSON == nil)
	cToDeserialize := &testResponseContainer{}
	require.NoError(t, json.Unmarshal(j, cToDeserialize))
	assert.Equal(t, "test string", cToDeserialize.TestString)
}

func TestWriteMiddlewareResponseHeadersAndBody(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		response := &http.Response{
			Header: http.Header{
				"Foo": []string{"foo"},
				"Grpc-Metadata-" + grpc.HTTPCodeMetadataKey: []string{"204"},
			},
		}
		container := defaultResponseContainer()
		responseJSON, err := json.Marshal(container)
		require.NoError(t, err)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		errJSON := WriteMiddlewareResponseHeadersAndBody(response, responseJSON, writer)
		require.Equal(t, true, errJSON == nil)
		v, ok := writer.Header()["Foo"]
		require.Equal(t, true, ok, "header not found")
		require.Equal(t, 1, len(v), "wrong number of header values")
		assert.Equal(t, "foo", v[0])
		v, ok = writer.Header()["Content-Length"]
		require.Equal(t, true, ok, "header not found")
		require.Equal(t, 1, len(v), "wrong number of header values")
		assert.Equal(t, "102", v[0])
		assert.Equal(t, 204, writer.Code)
		assert.DeepEqual(t, responseJSON, writer.Body.Bytes())
	})

	t.Run("GET_no_grpc_status_code_header", func(t *testing.T) {
		response := &http.Response{
			Header:     http.Header{},
			StatusCode: 204,
		}
		container := defaultResponseContainer()
		responseJSON, err := json.Marshal(container)
		require.NoError(t, err)
		writer := httptest.NewRecorder()

		errJSON := WriteMiddlewareResponseHeadersAndBody(response, responseJSON, writer)
		require.Equal(t, true, errJSON == nil)
		assert.Equal(t, 204, writer.Code)
	})

	t.Run("GET_invalid_status_code", func(t *testing.T) {
		response := &http.Response{
			Header: http.Header{},
		}

		// Set invalid status code.
		response.Header["Grpc-Metadata-"+grpc.HTTPCodeMetadataKey] = []string{"invalid"}

		container := defaultResponseContainer()
		responseJSON, err := json.Marshal(container)
		require.NoError(t, err)
		writer := httptest.NewRecorder()

		errJSON := WriteMiddlewareResponseHeadersAndBody(response, responseJSON, writer)
		require.Equal(t, false, errJSON == nil)
		assert.Equal(t, true, strings.Contains(errJSON.Msg(), "could not parse status code"))
		assert.Equal(t, http.StatusInternalServerError, errJSON.StatusCode())
	})

	t.Run("POST", func(t *testing.T) {
		response := &http.Response{
			Header:     http.Header{},
			StatusCode: 204,
		}
		container := defaultResponseContainer()
		responseJSON, err := json.Marshal(container)
		require.NoError(t, err)
		writer := httptest.NewRecorder()

		errJSON := WriteMiddlewareResponseHeadersAndBody(response, responseJSON, writer)
		require.Equal(t, true, errJSON == nil)
		assert.Equal(t, 204, writer.Code)
	})

	t.Run("POST_with_response_body", func(t *testing.T) {
		response := &http.Response{
			Header:     http.Header{},
			StatusCode: 204,
		}
		container := defaultResponseContainer()
		responseJSON, err := json.Marshal(container)
		require.NoError(t, err)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		errJSON := WriteMiddlewareResponseHeadersAndBody(response, responseJSON, writer)
		require.Equal(t, true, errJSON == nil)
		assert.Equal(t, 204, writer.Code)
		assert.DeepEqual(t, responseJSON, writer.Body.Bytes())
	})

	t.Run("POST_with_empty_json_body", func(t *testing.T) {
		response := &http.Response{
			Header:     http.Header{},
			StatusCode: 204,
		}
		responseJSON, err := json.Marshal(struct{}{})
		require.NoError(t, err)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		errJSON := WriteMiddlewareResponseHeadersAndBody(response, responseJSON, writer)
		require.Equal(t, true, errJSON == nil)
		assert.Equal(t, 204, writer.Code)
		assert.DeepEqual(t, []byte(nil), writer.Body.Bytes())
		assert.Equal(t, "0", writer.Header()["Content-Length"][0])
	})
}

func TestWriteError(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		responseHeader := http.Header{
			"Grpc-Metadata-" + grpc.CustomErrorMetadataKey: []string{"{\"CustomField\":\"bar\"}"},
		}
		errJSON := &testErrorJSON{
			Message: "foo",
			Code:    500,
		}
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		WriteError(writer, errJSON, responseHeader)
		v, ok := writer.Header()["Content-Length"]
		require.Equal(t, true, ok, "header not found")
		require.Equal(t, 1, len(v), "wrong number of header values")
		assert.Equal(t, "48", v[0])
		v, ok = writer.Header()["Content-Type"]
		require.Equal(t, true, ok, "header not found")
		require.Equal(t, 1, len(v), "wrong number of header values")
		assert.Equal(t, "application/json", v[0])
		assert.Equal(t, 500, writer.Code)
		eDeserialize := &testErrorJSON{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), eDeserialize))
		assert.Equal(t, "foo", eDeserialize.Message)
		assert.Equal(t, 500, eDeserialize.Code)
		assert.Equal(t, "bar", eDeserialize.CustomField)
	})

	t.Run("invalid_custom_error_header", func(t *testing.T) {
		logHook := test.NewGlobal()

		responseHeader := http.Header{
			"Grpc-Metadata-" + grpc.CustomErrorMetadataKey: []string{"invalid"},
		}

		WriteError(httptest.NewRecorder(), &testErrorJSON{}, responseHeader)
		assert.LogsContain(t, logHook, "Could not unmarshal custom error message")
	})
}
