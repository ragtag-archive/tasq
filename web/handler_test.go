package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/ragtag-archive/tasq/static"
	"github.com/ragtag-archive/tasq/web"
	"github.com/stretchr/testify/assert"
)

func getResponse[T any](t *testing.T, w *httptest.ResponseRecorder) (res web.Response[T]) {
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	return
}

func testList(t *testing.T, handler http.Handler, queueName string,
	xStatus int, xOk bool, xMsg string, xLen int) []string {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/"+queueName, nil)
	handler.ServeHTTP(w, r)
	listResponse := getResponse[web.ListResponse](t, w)
	assert.EqualValues(xStatus, w.Code)
	assert.EqualValues("application/json", w.Header().Get("Content-Type"))
	assert.EqualValues(xOk, listResponse.Ok)
	assert.EqualValues(xMsg, listResponse.Message)
	assert.EqualValues(xLen, len(listResponse.Payload.Tasks))
	assert.EqualValues(xLen, listResponse.Payload.Count)
	return listResponse.Payload.Tasks
}

func testPop(t *testing.T, handler http.Handler, queueName string,
	xStatus int, xOk bool, xMsg string) web.GetResponse {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/"+queueName, nil)
	handler.ServeHTTP(w, r)
	popResponse := getResponse[web.GetResponse](t, w)
	assert.EqualValues(xStatus, w.Code)
	assert.EqualValues("application/json", w.Header().Get("Content-Type"))
	assert.EqualValues(xOk, popResponse.Ok)
	assert.EqualValues(xMsg, popResponse.Message)
	return popResponse.Payload
}

func testInsert(t *testing.T, handler http.Handler, queueName string, data string,
	xStatus int, xOk bool, xMsg string) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/"+queueName, strings.NewReader(data))
	handler.ServeHTTP(w, r)
	insertResponse := getResponse[web.PutResponse](t, w)
	assert.EqualValues(xStatus, w.Code)
	assert.EqualValues("application/json", w.Header().Get("Content-Type"))
	assert.EqualValues(xOk, insertResponse.Ok)
	assert.EqualValues(xMsg, insertResponse.Message)
	assert.EqualValues(queueName+":"+data, insertResponse.Payload.Key)
}

func getHandler(t *testing.T) http.Handler {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return web.Handler(redisClient)
}

func TestBasic(t *testing.T) {
	assert := assert.New(t)
	handler := getHandler(t)

	// Test GET index
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)
	assert.EqualValues(200, w.Code)
	assert.EqualValues("text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	assert.EqualValues(static.IndexPage, w.Body.String())

	// Test list nonexistent queue
	testList(t, handler, "foo", 200, true, "", 0)

	// Test pop nonexistent queue
	testPop(t, handler, "foo", 404, false, web.MsgEmptyQueue)

	// Test insert
	testInsert(t, handler, "foo", "bar", 200, true, "")

	// Test list
	tasks := testList(t, handler, "foo", 200, true, "", 1)
	assert.EqualValues("foo:bar", tasks[0])

	// Shouldn't leak to other queues
	testList(t, handler, "bar", 200, true, "", 0)

	// Test pop
	popResponse := testPop(t, handler, "foo", 200, true, "")
	assert.EqualValues("foo:bar", popResponse.Key)
	assert.EqualValues("bar", popResponse.Data)

	// Test list again
	testList(t, handler, "foo", 200, true, "", 0)
}

func TestPriority(t *testing.T) {
	assert := assert.New(t)
	handler := getHandler(t)

	// Insert sample data
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for i := 0; i < len(items); i++ {
		for j := 0; j <= i; j++ {
			t.Logf("Inserting %s", items[i])
			testInsert(t, handler, "foo", items[i], 200, true, "")
		}
	}

	// Check length
	testList(t, handler, "foo", 200, true, "", len(items))

	// Pop all items
	for i := len(items) - 1; i >= 0; i-- {
		popResponse := testPop(t, handler, "foo", 200, true, "")
		assert.EqualValues(items[i], popResponse.Data)
	}

	// Check length
	testList(t, handler, "foo", 200, true, "", 0)

	// Insert in reverse order
	for i := len(items) - 1; i >= 0; i-- {
		for j := 0; j <= i; j++ {
			t.Logf("Inserting %s", items[i])
			testInsert(t, handler, "foo", items[i], 200, true, "")
		}
	}

	// Check length
	testList(t, handler, "foo", 200, true, "", len(items))

	// Pop all items
	for i := len(items) - 1; i >= 0; i-- {
		popResponse := testPop(t, handler, "foo", 200, true, "")
		assert.EqualValues(items[i], popResponse.Data)
	}

	// Check length
	testList(t, handler, "foo", 200, true, "", 0)
}
