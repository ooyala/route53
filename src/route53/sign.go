package route53

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

func (r53 *Route53) sign(hreq *http.Request) {
	r53.authLock.RLock()
	now := time.Now().UTC().Format(time.RFC1123)

	hash := hmac.New(sha256.New, []byte(r53.auth.SecretKey))
	hash.Write([]byte(now))

	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	header := fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s,", r53.auth.AccessKey)
	header += fmt.Sprintf("Algorithm=HmacSHA256,Signature=%s", signature)

	hreq.Header.Set("X-Amz-Date", now)
	hreq.Header.Set("X-Amzn-Authorization", header)
	hreq.Header.Set("X-Amz-Security-Token", r53.auth.Token())
	r53.authLock.RUnlock()
}
