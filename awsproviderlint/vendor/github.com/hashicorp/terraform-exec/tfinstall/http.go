package tfinstall

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	cleanhttp "github.com/hashicorp/go-cleanhttp"

	intversion "github.com/hashicorp/terraform-exec/internal/version"
)

type userAgentRoundTripper struct {
	inner     http.RoundTripper
	userAgent string
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", rt.userAgent)
	}
	return rt.inner.RoundTrip(req)
}

func newHTTPClient(appendUA string) *http.Client {
	appendUA = strings.TrimSpace(appendUA + " " + os.Getenv("TF_APPEND_USER_AGENT"))
	userAgent := strings.TrimSpace(fmt.Sprintf("HashiCorp-tfinstall/%s %s", intversion.ModuleVersion(), appendUA))

	cli := cleanhttp.DefaultPooledClient()
	cli.Transport = &userAgentRoundTripper{
		userAgent: userAgent,
		inner:     cli.Transport,
	}

	return cli
}
