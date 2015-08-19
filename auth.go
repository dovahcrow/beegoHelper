package beegoHelper

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"github.com/astaxie/beego/context"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var DigestAuthor = NewDigestAuthenticator()

func (this *Base) CheckBasicAuth(cred string, realm string) bool {
	this.Ctx.Output.Header("WWW-Authenticate", `Basic realm="`+realm+`"`)
	s := strings.SplitN(this.Ctx.Input.Header("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}
	if s[0] != "Basic" {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	user := strings.Split(cred, ":")
	return pair[0] == user[0] && pair[1] == user[1]
}

func (this *Base) CheckDigestAuth(cred string, realm string) bool {
	kv := strings.SplitN(cred, ":", 2)
	se := func(username, realm string) string {
		if username == kv[0] {
			return kv[1]
		}
		return ""
	}

	DigestAuthor.AddHeader(this.Ctx, realm)
	name, _ := DigestAuthor.CheckAuth(this.Ctx, se, realm)

	if name == kv[0] {
		return true
	}
	return false
}

type SecretProvider func(string, string) string

type digest_client struct {
	nc        uint64
	last_seen int64
}

type DigestAuth struct {
	Opaque           string
	PlainTextSecrets bool

	/*
	   Approximate size of Client's Cache. When actual number of
	   tracked client nonces exceeds
	   ClientCacheSize+ClientCacheTolerance, ClientCacheTolerance*2
	   older entries are purged.
	*/
	ClientCacheSize      int
	ClientCacheTolerance int

	clients map[string]*digest_client
	mutex   sync.Mutex
}

type digest_cache_entry struct {
	nonce     string
	last_seen int64
}

type digest_cache []digest_cache_entry

func (c digest_cache) Less(i, j int) bool {
	return c[i].last_seen < c[j].last_seen
}

func (c digest_cache) Len() int {
	return len(c)
}

func (c digest_cache) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

/*
 Remove count oldest entries from DigestAuth.clients
*/
func (a *DigestAuth) Purge(count int) {
	entries := make([]digest_cache_entry, 0, len(a.clients))
	for nonce, client := range a.clients {
		entries = append(entries, digest_cache_entry{nonce, client.last_seen})
	}
	cache := digest_cache(entries)
	sort.Sort(cache)
	for _, client := range cache[:count] {
		delete(a.clients, client.nonce)
	}
}

func (a *DigestAuth) AddHeader(ctx *context.Context, realm string) {
	if len(a.clients) > a.ClientCacheSize+a.ClientCacheTolerance {
		a.Purge(a.ClientCacheTolerance * 2)
	}
	nonce := RandomKey()
	a.clients[nonce] = &digest_client{nc: 0, last_seen: time.Now().UnixNano()}
	ctx.Output.Header("WWW-Authenticate",
		fmt.Sprintf(`Digest realm="%s", nonce="%s", opaque="%s", algorithm="MD5", qop="auth"`, realm, nonce, a.Opaque))
}

/*
 Parse Authorization header from the http.Request. Returns a map of
 auth parameters or nil if the header is not a valid parsable Digest
 auth header.
*/
func DigestAuthParams(c *context.Context) map[string]string {
	s := strings.SplitN(c.Input.Header("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Digest" {
		return nil
	}

	result := map[string]string{}
	for _, kv := range strings.Split(s[1], ",") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue

		}
		result[strings.Trim(parts[0], "\" ")] = strings.Trim(parts[1], "\" ")
	}
	return result
}

/*
 Check if request contains valid authentication data. Returns a pair
 of username, authinfo where username is the name of the authenticated
 user or an empty string and authinfo is the contents for the optional
 Authentication-Info response header.
*/
func (da *DigestAuth) CheckAuth(ctx *context.Context, se SecretProvider, realm string) (username string, authinfo *string) {
	da.mutex.Lock()
	defer da.mutex.Unlock()
	username = ""
	authinfo = nil
	auth := DigestAuthParams(ctx)
	if auth == nil || da.Opaque != auth["opaque"] || auth["algorithm"] != "MD5" || auth["qop"] != "auth" {
		return

	}

	// Check if the requested URI matches auth header
	switch u, err := url.Parse(auth["uri"]); {
	case err != nil:
		return
	case ctx.Input.Request.URL == nil:
		return
	case len(u.Path) > len(ctx.Input.Request.URL.Path):
		return
	case !strings.HasPrefix(ctx.Input.Request.URL.Path, u.Path):
		return
	}

	HA1 := se(auth["username"], realm)
	HA1 = H(auth["username"] + ":" + realm + ":" + HA1)
	HA2 := H(ctx.Input.Request.Method + ":" + auth["uri"])
	KD := H(strings.Join([]string{HA1, auth["nonce"], auth["nc"], auth["cnonce"], auth["qop"], HA2}, ":"))
	if subtle.ConstantTimeCompare([]byte(KD), []byte(auth["response"])) != 1 {
		return
	}

	// At this point crypto checks are completed and validated.
	// Now check if the session is valid.

	nc, err := strconv.ParseUint(auth["nc"], 16, 64)
	if err != nil {
		return
	}
	if client, ok := da.clients[auth["nonce"]]; !ok {
		return
	} else {
		if client.nc != 0 && client.nc >= nc {
			return
		}
		client.nc = nc
		client.last_seen = time.Now().UnixNano()
	}
	resp_HA2 := H(":" + auth["uri"])
	rspauth := H(strings.Join([]string{HA1, auth["nonce"], auth["nc"], auth["cnonce"], auth["qop"], resp_HA2}, ":"))
	info := fmt.Sprintf(`qop="auth", rspauth="%s", cnonce="%s", nc="%s"`, rspauth, auth["cnonce"], auth["nc"])
	return auth["username"], &info
}

/*
 Default values for ClientCacheSize and ClientCacheTolerance for DigestAuth
*/
const DefaultClientCacheSize = 1000
const DefaultClientCacheTolerance = 100

func NewDigestAuthenticator() *DigestAuth {
	da := &DigestAuth{
		Opaque:               RandomKey(),
		PlainTextSecrets:     false,
		ClientCacheSize:      DefaultClientCacheSize,
		ClientCacheTolerance: DefaultClientCacheTolerance,
		clients:              map[string]*digest_client{},
	}
	return da
}

/*
 Return a random 16-byte base64 alphabet string
*/
func RandomKey() string {
	k := make([]byte, 12)
	for bytes := 0; bytes < len(k); {
		n, err := rand.Read(k[bytes:])
		if err != nil {
			panic("rand.Read() failed")

		}
		bytes += n

	}
	return base64.StdEncoding.EncodeToString(k)
}

/*
 H function for MD5 algorithm (returns a lower-case hex MD5 digest)
*/
func H(data string) string {
	digest := md5.New()
	digest.Write([]byte(data))
	return fmt.Sprintf("%x", digest.Sum(nil))
}
