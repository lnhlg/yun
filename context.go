package yun

import (
	"net/http"
	"path/filepath"
	"io"
	"encoding/json"
	"encoding/xml"
	"mime"
	"io/ioutil"
	"strconv"
	"bytes"
	"errors"
	"reflect"
	"unsafe"
)

type (
	Context struct {
		tempwriter	responseWriter
		request 	*http.Request
		ResponseWriter
		Params   	Params
		handlers	Handlers
		index		int16
		hcount		int16

		keys     map[string]interface{}
	}
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

const (
	charsetUTF8 = "charset=utf-8"
	outside = 255
)

// Headers
const (
	HeaderAcceptEncoding                = "Accept-Encoding"
	HeaderAllow                         = "Allow"
	HeaderAuthorization                 = "Authorization"
	HeaderContentDisposition            = "Content-Disposition"
	HeaderContentEncoding               = "Content-Encoding"
	HeaderContentLength                 = "Content-Length"
	HeaderContentType                   = "Content-Type"
	HeaderCookie                        = "Cookie"
	HeaderSetCookie                     = "Set-Cookie"
	HeaderIfModifiedSince               = "If-Modified-Since"
	HeaderLastModified                  = "Last-Modified"
	HeaderLocation                      = "Location"
	HeaderUpgrade                       = "Upgrade"
	HeaderVary                          = "Vary"
	HeaderWWWAuthenticate               = "WWW-Authenticate"
	HeaderXForwardedProto               = "X-Forwarded-Proto"
	HeaderXHTTPMethodOverride           = "X-HTTP-Method-Override"
	HeaderXForwardedFor                 = "X-Forwarded-For"
	HeaderXRealIP                       = "X-Real-IP"
	HeaderServer                        = "Server"
	HeaderOrigin                        = "Origin"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

func (c *Context) reset(w http.ResponseWriter, req *http.Request) {
	c.tempwriter.reset(w)
	c.ResponseWriter = &c.tempwriter
	c.request = req
	c.keys = nil
	c.index = -1
	c.Params = c.Params[0:0]
	c.handlers = nil
	c.hcount = 0
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) Response() http.ResponseWriter {
	return c.ResponseWriter
}

func (c *Context) Next() {
	c.index++
	if c.index < c.hcount {
		c.handlers[c.index](c)
	}
}

func (c *Context) HTML(code int, html string) (err error) {
	c.Response().Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	c.WriteHeader(code)
	_, err = c.Write(c.StringToBytes(html))
	return
}

func (c *Context) String(code int, s string) (err error) {
	c.Response().Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	c.WriteHeader(code)
	_, err = c.Write(c.StringToBytes(s))
	return
}

func (c *Context) JSON(code int, i interface{}) (err error) {
	b, err := json.Marshal(i)

	if err != nil {
		return err
	}

	return c.JSONBlob(code, b)
}

func (c *Context) JSONBlob(code int, b []byte) (err error) {
	c.Response().Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	c.WriteHeader(code)
	_, err = c.Write(b)
	return
}

func (c *Context) JSONP(code int, callback string, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c.Response().Header().Set(HeaderContentType, MIMEApplicationJavaScriptCharsetUTF8)
	c.WriteHeader(code)
	if _, err = c.Response().Write(c.StringToBytes(callback + "(")); err != nil {
		return
	}
	if _, err = c.Response().Write(b); err != nil {
		return
	}
	_, err = c.Response().Write([]byte(");"))
	return
}

func (c *Context) XML(code int, i interface{}) (err error) {
	b, err := xml.Marshal(i)

	if err != nil {
		return err
	}

	return c.XMLBlob(code, b)
}

func (c *Context) XMLBlob(code int, b []byte) (err error) {
	c.Response().Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	c.WriteHeader(code)
	if _, err = c.Write(c.StringToBytes(xml.Header)); err != nil {
		return
	}
	_, err = c.Write(b)
	return
}

func (c *Context) File(file string) {
	http.ServeFile(c.Response(), c.Request(), file)
}

func (c *Context) Attachment(r io.ReadSeeker, name string) (err error) {
	c.Response().Header().Set(HeaderContentType, ContentTypeByExtension(name))
	c.Response().Header().Set(HeaderContentDisposition, "attachment; filename="+name)
	c.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.ResponseWriter, r)
	return
}

func (c *Context) NoContent(code int) error {
	c.WriteHeader(code)
	return nil
}

func ContentTypeByExtension(name string) (t string) {
	if t = mime.TypeByExtension(filepath.Ext(name)); t == "" {
		t = MIMEOctetStream
	}
	return
}

func (c *Context) Set(key string, value interface{}) {
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}

	c.keys[key] = value
}

func (c *Context) Get(key string) (value interface{}, exists bool) {
	if c.keys != nil {
		value, exists = c.keys[key]
	}
	return
}

func (c *Context) Send(method, url string, body []byte) ([]byte, int, error) {
	client := new(http.Client)
	b := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, url, b)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return res, resp.StatusCode, nil
}

func (c *Context) Form(key string) string {
	if v, ok := c.getForm(key); ok {
		return v
	}

	return ""
}

func (c *Context) MustForm(key string) (string, error) {
	if v, ok := c.getForm(key); ok {
		return v, nil
	}

	return "", errors.New("Query Paramete \"" + key + "\" does not exist")
}

func (c *Context) FormInt(key string) int {
	if v, ok := c.getForm(key); ok {
		return strconv.Atoi(v)
	}
	return 0
}

func (c *Context) MustFormInt(key string) (int, error) {
	if v, ok := c.getForm(key); ok {
		return strconv.Atoi(v), nil
	}
	return 0, errors.New("Query Paramete \"" + key + "\" does not exist")
}

func (c *Context) Body() ([]byte, error) {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}
	defer c.Request().Body.Close()

	return body, nil
}

func (c *Context) DecodeJson(obj interface{}) error {
	body, err := c.Body()
	if err != nil {
		return err
	}

	return json.Unmarshal(body, obj)
}

func (c *Context)BytesToString(b []byte) string {
	byteshead := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	strhead := reflect.StringHeader{byteshead.Data, byteshead.Len}
	return *(*string)(unsafe.Pointer(&strhead))
}

func (c *Context)StringToBytes(s string) []byte {
	strhead := (*reflect.StringHeader)(unsafe.Pointer(&s))
	byteshead := reflect.SliceHeader{strhead.Data, strhead.Len, 0}
	return *(*[]byte)(unsafe.Pointer(&byteshead))
}

func (c *Context) IsAborted() bool {
	return c.index >= outside
}

func (c *Context) Abort() {
	c.index = outside
}

func (c *Context) AbortCode(code int) {
	c.WriteHeader(code)
	c.Abort()
}

func (c *Context) getForm(key string) (string, bool) {
	if values, ok := c.Request().URL.Query()[key]; ok && len(values) > 0 {
		return values[0], true
	}
	return "", false
}

func (c *Context) setHandlers(handlers Handlers) {
	c.handlers = handlers
	c.hcount = int16(len(handlers))
}
