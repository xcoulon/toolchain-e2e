package wait

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/common/expfmt"
)

func getCounter(svc string, family string, labelKey string, labelValue string) (float64, error) {
	client := http.Client{
		Timeout: time.Duration(1 * time.Second),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Get(fmt.Sprintf("%s/metrics", svc))
	if err != nil {
		return -1, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("unexpected response status code: '%d'", resp.StatusCode)
	}
	metrics, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	// parse the metrics
	parser := expfmt.TextParser{}
	families, err := parser.TextToMetricFamilies(bytes.NewReader(metrics))
	if err != nil {
		if err != nil {
			return -1, err
		}
	}
	for _, f := range families {
		if f.GetName() == family {
			for _, m := range f.GetMetric() {
				for _, l := range m.GetLabel() {
					// fmt.Printf("family_name: '%s' / label_name: '%s' / label_value: '%s'\n", name, l.GetName(), l.GetValue())
					if l.GetName() == labelKey && l.GetValue() == labelValue {
						return m.GetCounter().GetValue(), nil
					}
				}
			}
		}
	}
	return -1, nil
}
