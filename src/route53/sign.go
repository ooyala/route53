package route53

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"net/http"
	"time"
)

func sign(auth aws.Auth, hreq *http.Request) {
	now := time.Now().UTC().Format(time.RFC1123)

	hash := hmac.New(sha256.New, []byte(auth.SecretKey))
	hash.Write([]byte(now))

	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	header := fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s,", auth.AccessKey)
	header += fmt.Sprintf("Algorithm=HmacSHA256,Signature=%s", signature)

	hreq.Header.Set("X-Amz-Date", now)
	hreq.Header.Set("X-Amzn-Authorization", header)
	hreq.Header.Set("X-Amz-Security-Token", auth.Token())
}
