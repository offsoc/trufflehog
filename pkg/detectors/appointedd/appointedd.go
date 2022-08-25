package appointedd

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"io/ioutil"
	
	"github.com/trufflesecurity/trufflehog/v3/pkg/common"
	"github.com/trufflesecurity/trufflehog/v3/pkg/detectors"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/detectorspb"
)

type Scanner struct{}

// Ensure the Scanner satisfies the interface at compile time
var _ detectors.Detector = (*Scanner)(nil)

var (
	client = common.SaneHttpClient()

	//Make sure that your group is surrounded in boundry characters such as below to reduce false positives
	keyPat = regexp.MustCompile(detectors.PrefixRegex([]string{"appointedd"}) + `\b([a-zA-Z0-9=+]{88})`)
)

// Keywords are used for efficiently pre-filtering chunks.
// Use identifiers in the secret preferably, or the provider name.
func (s Scanner) Keywords() []string {
	return []string{"appointedd"}
}

// FromData will find and optionally verify appointedd secrets in a given set of bytes.
func (s Scanner) FromData(ctx context.Context, verify bool, data []byte) (results []detectors.Result, err error) {
	dataStr := string(data)

	matches := keyPat.FindAllStringSubmatch(dataStr, -1)

	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		resMatch := strings.TrimSpace(match[1])

		s1 := detectors.Result{
			DetectorType: detectorspb.DetectorType_Appointedd,
			Raw:          []byte(resMatch),
		}
		if verify {
			req, err := http.NewRequestWithContext(ctx, "GET", "https://api.appointedd.com/v1/availability/slots", nil)
			if err != nil {
				continue
			}
			req.Header.Add("X-API-KEY", resMatch)
			res, err := client.Do(req)
			if err == nil {
				defer res.Body.Close()
				bodyBytes, err := ioutil.ReadAll(res.Body)
				if err != nil {
					continue
				}
				body := string(bodyBytes)

				if strings.Contains(body, "total") {
					s1.Verified = true
				} else {
					if detectors.IsKnownFalsePositive(resMatch, detectors.DefaultFalsePositives, true) {
						continue
					}
				}
			}
		}

		results = append(results, s1)
	}

	return detectors.CleanResults(results), nil
}
