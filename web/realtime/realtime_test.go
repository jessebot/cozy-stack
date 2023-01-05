package realtime

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cozy/cozy-stack/model/instance"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/realtime"
	"github.com/cozy/cozy-stack/tests/testutils"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ts *httptest.Server
var inst *instance.Instance
var token string

type testDoc struct {
	id      string
	doctype string
}

func TestRealtime(t *testing.T) {
	if testing.Short() {
		t.Skip("an instance is required for this test: test skipped due to the use of --short flag")
	}

	config.UseTestFile()
	testutils.NeedCouchdb()
	setup := testutils.NewSetup(nil, t.Name())
	t.Cleanup(setup.Cleanup)
	inst = setup.GetTestInstance()
	_, token = setup.GetTestClient("io.cozy.foos io.cozy.bars io.cozy.bazs")
	ts = setup.GetTestServer("/realtime", Routes)

	t.Run("WSNoAuth", func(t *testing.T) {
		u := strings.Replace(ts.URL+"/realtime/", "http", "ws", 1)
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NoError(t, err)
		defer ws.Close()

		msg := `{"method": "SUBSCRIBE", "payload": { "type": "io.cozy.foos" }}`
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
		assert.NoError(t, err)

		var res map[string]interface{}
		err = ws.ReadJSON(&res)
		assert.NoError(t, err)
		assert.Equal(t, "error", res["event"])
		payload := res["payload"].(map[string]interface{})
		assert.Equal(t, "405 Method Not Allowed", payload["status"])
		assert.Equal(t, "method not allowed", payload["code"])
		assert.Equal(t, "The SUBSCRIBE method is not supported", payload["title"])
	})

	t.Run("WSInvalidToken", func(t *testing.T) {
		u := strings.Replace(ts.URL+"/realtime/", "http", "ws", 1)
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NoError(t, err)
		defer ws.Close()

		auth := `{"method": "AUTH", "payload": "123456789"}`
		err = ws.WriteMessage(websocket.TextMessage, []byte(auth))
		assert.NoError(t, err)

		var res map[string]interface{}
		err = ws.ReadJSON(&res)
		assert.NoError(t, err)
		assert.Equal(t, "error", res["event"])
		payload := res["payload"].(map[string]interface{})
		assert.Equal(t, "401 Unauthorized", payload["status"])
		assert.Equal(t, "unauthorized", payload["code"])
		assert.Equal(t, "The authentication has failed", payload["title"])
	})

	t.Run("WSNoPermissionsForADoctype", func(t *testing.T) {
		u := strings.Replace(ts.URL+"/realtime/", "http", "ws", 1)
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		assert.NoError(t, err)
		defer ws.Close()

		auth := fmt.Sprintf(`{"method": "AUTH", "payload": "%s"}`, token)
		err = ws.WriteMessage(websocket.TextMessage, []byte(auth))
		assert.NoError(t, err)

		msg := `{"method": "SUBSCRIBE", "payload": { "type": "io.cozy.contacts" }}`
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
		assert.NoError(t, err)

		var res map[string]interface{}
		err = ws.ReadJSON(&res)
		assert.NoError(t, err)
		assert.Equal(t, "error", res["event"])
		payload := res["payload"].(map[string]interface{})
		assert.Equal(t, "403 Forbidden", payload["status"])
		assert.Equal(t, "forbidden", payload["code"])
		assert.Equal(t, "The application can't subscribe to io.cozy.contacts", payload["title"])
	})

	t.Run("WSSuccess", func(t *testing.T) {
		u := strings.Replace(ts.URL+"/realtime/", "http", "ws", 1)
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		require.NoError(t, err)

		defer ws.Close()

		auth := fmt.Sprintf(`{"method": "AUTH", "payload": "%s"}`, token)
		err = ws.WriteMessage(websocket.TextMessage, []byte(auth))
		require.NoError(t, err)

		msg := `{"method": "SUBSCRIBE", "payload": { "type": "io.cozy.foos" }}`
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
		require.NoError(t, err)

		msg = `{"method": "SUBSCRIBE", "payload": { "type": "io.cozy.bars", "id": "bar-one" }}`
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
		require.NoError(t, err)

		msg = `{"method": "SUBSCRIBE", "payload": { "type": "io.cozy.bars", "id": "bar-two" }}`
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
		require.NoError(t, err)

		h := realtime.GetHub()
		var res map[string]interface{}
		time.Sleep(30 * time.Millisecond)

		h.Publish(inst, realtime.EventUpdate, &testDoc{
			doctype: "io.cozy.foos",
			id:      "foo-one",
		}, nil)
		err = ws.ReadJSON(&res)
		assert.NoError(t, err)
		assert.Equal(t, "UPDATED", res["event"])
		payload := res["payload"].(map[string]interface{})
		assert.Equal(t, "io.cozy.foos", payload["type"])
		assert.Equal(t, "foo-one", payload["id"])

		h.Publish(inst, realtime.EventDelete, &testDoc{
			doctype: "io.cozy.foos",
			id:      "foo-two",
		}, nil)
		err = ws.ReadJSON(&res)
		assert.NoError(t, err)
		assert.Equal(t, "DELETED", res["event"])
		payload = res["payload"].(map[string]interface{})
		assert.Equal(t, "io.cozy.foos", payload["type"])
		assert.Equal(t, "foo-two", payload["id"])

		h.Publish(inst, realtime.EventCreate, &testDoc{
			doctype: "io.cozy.bars",
			id:      "bar-three",
		}, nil)
		// No event

		h.Publish(inst, realtime.EventCreate, &testDoc{
			doctype: "io.cozy.bars",
			id:      "bar-one",
		}, nil)
		err = ws.ReadJSON(&res)
		assert.NoError(t, err)
		assert.Equal(t, "CREATED", res["event"])
		payload = res["payload"].(map[string]interface{})
		assert.Equal(t, "io.cozy.bars", payload["type"])
		assert.Equal(t, "bar-one", payload["id"])

		msg = `{"method": "UNSUBSCRIBE", "payload": { "type": "io.cozy.bars", "id": "bar-one" }}`
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
		require.NoError(t, err)

		time.Sleep(30 * time.Millisecond)

		h.Publish(inst, realtime.EventUpdate, &testDoc{
			doctype: "io.cozy.bars",
			id:      "bar-one",
		}, nil)
		// No event

		h.Publish(inst, realtime.EventUpdate, &testDoc{
			doctype: "io.cozy.bars",
			id:      "bar-two",
		}, nil)
		err = ws.ReadJSON(&res)
		assert.NoError(t, err)
		assert.Equal(t, "UPDATED", res["event"])
		payload = res["payload"].(map[string]interface{})
		assert.Equal(t, "io.cozy.bars", payload["type"])
		assert.Equal(t, "bar-two", payload["id"])
	})

	t.Run("WSNotify", func(t *testing.T) {
		u := strings.Replace(ts.URL+"/realtime/", "http", "ws", 1)
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		require.NoError(t, err)

		defer ws.Close()

		auth := fmt.Sprintf(`{"method": "AUTH", "payload": "%s"}`, token)
		err = ws.WriteMessage(websocket.TextMessage, []byte(auth))
		require.NoError(t, err)

		msg := `{"method": "SUBSCRIBE", "payload": { "type": "io.cozy.bazs", "id": "baz-one" }}`
		err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
		require.NoError(t, err)

		time.Sleep(30 * time.Millisecond)
		body := `{"hello": "world"}`
		req, _ := http.NewRequest("POST", ts.URL+"/realtime/io.cozy.bazs/baz-one", bytes.NewBufferString(body))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)
		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		var resp map[string]interface{}
		err = ws.ReadJSON(&resp)
		require.NoError(t, err)

		assert.Equal(t, "NOTIFIED", resp["event"])
		payload := resp["payload"].(map[string]interface{})
		assert.Equal(t, "io.cozy.bazs", payload["type"])
		assert.Equal(t, "baz-one", payload["id"])
		doc := payload["doc"].(map[string]interface{})
		assert.Equal(t, "world", doc["hello"])
	})
}

func (t *testDoc) ID() string      { return t.id }
func (t *testDoc) DocType() string { return t.doctype }
func (t *testDoc) MarshalJSON() ([]byte, error) {
	j := `{"_id":"` + t.id + `", "_type":"` + t.doctype + `"}`
	return []byte(j), nil
}
