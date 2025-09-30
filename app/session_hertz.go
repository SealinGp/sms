package app

import (
	"context"
	"github.com/Akvicor/glog"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"sms/config"
)

func sessionVerify(ctx context.Context, c *app.RequestContext) bool {
	// Convert Hertz request context to standard HTTP request for session compatibility
	req := &http.Request{
		Header: make(http.Header),
	}

	// Copy relevant headers
	c.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Get session
	session, err := SessionStore.Get(req, config.Global.Session.Name)
	if err != nil {
		glog.Debug("Session error: %v", err)
		return false
	}

	username, ok := session.Values["username"]
	if !ok {
		glog.Debug("No username in session")
		return false
	}

	usernameStr, ok := username.(string)
	if !ok {
		glog.Debug("Username is not string")
		return false
	}

	return usernameStr == config.Global.Security.Username
}

func sessionUpdate(ctx context.Context, c *app.RequestContext, username string) {
	// Create a mock response writer to work with gorilla sessions
	mockWriter := &mockResponseWriter{
		header: make(http.Header),
	}

	// Convert Hertz request context to standard HTTP request
	req := &http.Request{
		Header: make(http.Header),
	}

	// Copy relevant headers
	c.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Get session
	session, err := SessionStore.Get(req, config.Global.Session.Name)
	if err != nil {
		glog.Error("Session error: %v", err)
		return
	}

	// Update session
	session.Values["username"] = username
	err = session.Save(req, mockWriter)
	if err != nil {
		glog.Error("Failed to save session: %v", err)
		return
	}

	// Copy session cookies to Hertz response
	for _, cookie := range mockWriter.header["Set-Cookie"] {
		c.Response.Header.Set("Set-Cookie", cookie)
	}
}

func keyVerify(ctx context.Context, c *app.RequestContext) bool {
	key := string(c.Query("key"))
	if key == "" {
		key = string(c.PostForm("key"))
	}
	return key == config.Global.Security.AccessKey
}

// mockResponseWriter implements http.ResponseWriter for session compatibility
type mockResponseWriter struct {
	header http.Header
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	// Do nothing
}