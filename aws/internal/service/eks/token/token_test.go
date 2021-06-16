/*
This file is a hard copy of:
https://github.com/kubernetes-sigs/aws-iam-authenticator/blob/7547c74e660f8d34d9980f2c69aa008eed1f48d0/pkg/token/token_test.go

With the following modifications:

 - Fix staticcheck reports
 - Ignore errorlint reports
 - Refactor deprecated io/ioutil in Go 1.16
*/

package token

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func validationErrorTest(t *testing.T, token string, expectedErr string) {
	t.Helper()
	_, err := tokenVerifier{}.Verify(token)
	errorContains(t, err, expectedErr)
}

func validationSuccessTest(t *testing.T, token string) {
	t.Helper()
	arn := "arn:aws:iam::123456789012:user/Alice"
	account := "123456789012"
	userID := "Alice"
	_, err := newVerifier(200, jsonResponse(arn, account, userID), nil).Verify(token)
	if err != nil {
		t.Errorf("received unexpected error: %s", err)
	}
}

func errorContains(t *testing.T, err error, expectedErr string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("err should have contained '%s' was '%s'", expectedErr, err)
	}
}

func assertSTSError(t *testing.T, err error) {
	t.Helper()
	if _, ok := err.(STSError); !ok { // nolint:errorlint
		t.Errorf("Expected err %v to be an STSError but was not", err)
	}
}

var (
	now        = time.Now()
	timeStr    = now.UTC().Format("20060102T150405Z")
	validURL   = fmt.Sprintf("https://sts.amazonaws.com/?action=GetCallerIdentity&X-Amz-Credential=ASIABCDEFGHIJKLMNOPQ%%2F20191216%%2Fus-west-2%%2Fs3%%2Faws4_request&x-amz-signedheaders=x-k8s-aws-id&x-amz-expires=60&x-amz-date=%s", timeStr)
	validToken = toToken(validURL)
)

func toToken(url string) string {
	return v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(url))
}

func newVerifier(statusCode int, body string, err error) Verifier {
	var rc io.ReadCloser
	if body != "" {
		rc = io.NopCloser(bytes.NewReader([]byte(body)))
	}
	return tokenVerifier{
		client: &http.Client{
			Transport: &roundTripper{
				err: err,
				resp: &http.Response{
					StatusCode: statusCode,
					Body:       rc,
				},
			},
		},
	}
}

type roundTripper struct {
	err  error
	resp *http.Response
}

type errorReadCloser struct {
}

func (r errorReadCloser) Read(b []byte) (int, error) {
	return 0, errors.New("An Error")
}

func (r errorReadCloser) Close() error {
	return nil
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.resp, rt.err
}

func jsonResponse(arn, account, userid string) string {
	response := getCallerIdentityWrapper{}
	response.GetCallerIdentityResponse.GetCallerIdentityResult.Account = account
	response.GetCallerIdentityResponse.GetCallerIdentityResult.Arn = arn
	response.GetCallerIdentityResponse.GetCallerIdentityResult.UserID = userid
	data, _ := json.Marshal(response)
	return string(data)
}

func TestSTSEndpoints(t *testing.T) {
	verifier := tokenVerifier{}
	chinaR := "sts.amazonaws.com.cn"
	globalR := "sts.amazonaws.com"
	usEast1R := "sts.us-east-1.amazonaws.com"
	usEast2R := "sts.us-east-2.amazonaws.com"
	usWest1R := "sts.us-west-1.amazonaws.com"
	usWest2R := "sts.us-west-2.amazonaws.com"
	apSouth1R := "sts.ap-south-1.amazonaws.com"
	apNorthEast1R := "sts.ap-northeast-1.amazonaws.com"
	apNorthEast2R := "sts.ap-northeast-2.amazonaws.com"
	apSouthEast1R := "sts.ap-southeast-1.amazonaws.com"
	apSouthEast2R := "sts.ap-southeast-2.amazonaws.com"
	caCentral1R := "sts.ca-central-1.amazonaws.com"
	euCenteral1R := "sts.eu-central-1.amazonaws.com"
	euWest1R := "sts.eu-west-1.amazonaws.com"
	euWest2R := "sts.eu-west-2.amazonaws.com"
	euWest3R := "sts.eu-west-3.amazonaws.com"
	euNorth1R := "sts.eu-north-1.amazonaws.com"
	saEast1R := "sts.sa-east-1.amazonaws.com"

	hosts := []string{chinaR, globalR, usEast1R, usEast2R, usWest1R, usWest2R, apSouth1R, apNorthEast1R, apNorthEast2R, apSouthEast1R, apSouthEast2R, caCentral1R, euCenteral1R, euWest1R, euWest2R, euWest3R, euNorth1R, saEast1R}

	for _, host := range hosts {
		if err := verifier.verifyHost(host); err != nil {
			t.Errorf("%s is not valid endpoints host", host)
		}
	}
}

func TestVerifyTokenPreSTSValidations(t *testing.T) {
	b := make([]byte, maxTokenLenBytes+1)
	s := string(b)
	validationErrorTest(t, s, "token is too large")
	validationErrorTest(t, "k8s-aws-v2.asdfasdfa", "token is missing expected \"k8s-aws-v1.\" prefix")
	validationErrorTest(t, "k8s-aws-v1.decodingerror", "illegal base64 data")
	validationErrorTest(t, toToken(":ab:cd.af:/asda"), "missing protocol scheme")
	validationErrorTest(t, toToken("http://"), "unexpected scheme")
	validationErrorTest(t, toToken("https://google.com"), fmt.Sprintf("unexpected hostname %q in pre-signed URL", "google.com"))
	validationErrorTest(t, toToken("https://sts.cn-north-1.amazonaws.com.cn/abc"), "unexpected path in pre-signed URL")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/abc"), "unexpected path in pre-signed URL")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/?NoInWhiteList=abc"), "non-whitelisted query parameter")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/?action=get&action=post"), "query parameter with multiple values not supported")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/?action=NotGetCallerIdenity"), "unexpected action parameter in pre-signed URL")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=abc%3bx-k8s-aws-i%3bdef"), "client did not sign the x-k8s-aws-id header in the pre-signed URL")
	validationErrorTest(t, toToken(fmt.Sprintf("https://sts.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-date=%s&x-amz-expires=9999999", timeStr)), "invalid X-Amz-Expires parameter in pre-signed URL")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-date=xxxxxxx&x-amz-expires=60"), "error parsing X-Amz-Date parameter")
	validationErrorTest(t, toToken("https://sts.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-date=19900422T010203Z&x-amz-expires=60"), "X-Amz-Date parameter is expired")
	validationSuccessTest(t, toToken(fmt.Sprintf("https://sts.us-east-2.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-date=%s&x-amz-expires=60", timeStr)))
	validationSuccessTest(t, toToken(fmt.Sprintf("https://sts.ap-northeast-2.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-date=%s&x-amz-expires=60", timeStr)))
	validationSuccessTest(t, toToken(fmt.Sprintf("https://sts.ca-central-1.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-date=%s&x-amz-expires=60", timeStr)))
	validationSuccessTest(t, toToken(fmt.Sprintf("https://sts.eu-west-1.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-date=%s&x-amz-expires=60", timeStr)))
	validationSuccessTest(t, toToken(fmt.Sprintf("https://sts.sa-east-1.amazonaws.com/?action=GetCallerIdentity&x-amz-signedheaders=x-k8s-aws-id&x-amz-date=%s&x-amz-expires=60", timeStr)))
}

func TestVerifyHTTPError(t *testing.T) {
	_, err := newVerifier(0, "", errors.New("an error")).Verify(validToken)
	errorContains(t, err, "error during GET: an error")
	assertSTSError(t, err)
}

func TestVerifyHTTP403(t *testing.T) {
	_, err := newVerifier(403, " ", nil).Verify(validToken)
	errorContains(t, err, "error from AWS (expected 200, got")
	assertSTSError(t, err)
}

func TestVerifyBodyReadError(t *testing.T) {
	verifier := tokenVerifier{
		client: &http.Client{
			Transport: &roundTripper{
				err: nil,
				resp: &http.Response{
					StatusCode: 200,
					Body:       errorReadCloser{},
				},
			},
		},
	}
	_, err := verifier.Verify(validToken)
	errorContains(t, err, "error reading HTTP result")
	assertSTSError(t, err)
}

func TestVerifyUnmarshalJSONError(t *testing.T) {
	_, err := newVerifier(200, "xxxx", nil).Verify(validToken)
	errorContains(t, err, "invalid character")
	assertSTSError(t, err)
}

func TestVerifyInvalidCanonicalARNError(t *testing.T) {
	_, err := newVerifier(200, jsonResponse("arn", "1000", "userid"), nil).Verify(validToken)
	errorContains(t, err, "arn 'arn' is invalid:")
	assertSTSError(t, err)
}

func TestVerifyInvalidUserIDError(t *testing.T) {
	_, err := newVerifier(200, jsonResponse("arn:aws:iam::123456789012:user/Alice", "123456789012", "not:vailid:userid"), nil).Verify(validToken)
	errorContains(t, err, "malformed UserID")
	assertSTSError(t, err)
}

func TestVerifyNoSession(t *testing.T) {
	arn := "arn:aws:iam::123456789012:user/Alice"
	account := "123456789012"
	userID := "Alice"
	accessKeyID := "ASIABCDEFGHIJKLMNOPQ"
	identity, err := newVerifier(200, jsonResponse(arn, account, userID), nil).Verify(validToken)
	if err != nil {
		t.Errorf("expected error to be nil was %q", err)
	}
	if identity.AccessKeyID != accessKeyID {
		t.Errorf("expected AccessKeyID to be %q but was %q", accessKeyID, identity.AccessKeyID)
	}
	if identity.ARN != arn {
		t.Errorf("expected ARN to be %q but was %q", arn, identity.ARN)
	}
	if identity.CanonicalARN != arn {
		t.Errorf("expected CanonicalARN to be %q but was %q", arn, identity.CanonicalARN)
	}
	if identity.UserID != userID {
		t.Errorf("expected Username to be %q but was %q", userID, identity.UserID)
	}
}

func TestVerifySessionName(t *testing.T) {
	arn := "arn:aws:iam::123456789012:user/Alice"
	account := "123456789012"
	userID := "Alice"
	session := "session-name"
	identity, err := newVerifier(200, jsonResponse(arn, account, userID+":"+session), nil).Verify(validToken)
	if err != nil {
		t.Errorf("expected error to be nil was %q", err)
	}
	if identity.UserID != userID {
		t.Errorf("expected Username to be %q but was %q", userID, identity.UserID)
	}
	if identity.SessionName != session {
		t.Errorf("expected Session to be %q but was %q", session, identity.SessionName)
	}
}

func TestVerifyCanonicalARN(t *testing.T) {
	arn := "arn:aws:sts::123456789012:assumed-role/Alice/extra"
	canonicalARN := "arn:aws:iam::123456789012:role/Alice"
	account := "123456789012"
	userID := "Alice"
	session := "session-name"
	identity, err := newVerifier(200, jsonResponse(arn, account, userID+":"+session), nil).Verify(validToken)
	if err != nil {
		t.Errorf("expected error to be nil was %q", err)
	}
	if identity.ARN != arn {
		t.Errorf("expected ARN to be %q but was %q", arn, identity.ARN)
	}
	if identity.CanonicalARN != canonicalARN {
		t.Errorf("expected CannonicalARN to be %q but was %q", canonicalARN, identity.CanonicalARN)
	}
}
