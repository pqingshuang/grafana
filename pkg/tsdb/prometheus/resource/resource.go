package resource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana/pkg/infra/httpclient"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/infra/tracing"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb/intervalv2"
	"github.com/grafana/grafana/pkg/tsdb/prometheus/client"
	"github.com/grafana/grafana/pkg/util/maputil"
)

type Resource struct {
	intervalCalculator intervalv2.Calculator
	tracer             tracing.Tracer
	provider           *client.Provider
	log                log.Logger
	ID                 int64
	URL                string
	TimeInterval       string
}

func New(
	httpClientProvider httpclient.Provider,
	cfg *setting.Cfg,
	features featuremgmt.FeatureToggles,
	tracer tracing.Tracer,
	settings backend.DataSourceInstanceSettings,
	plog log.Logger,
) (*Resource, error) {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(settings.JSONData, &jsonData); err != nil {
		return nil, fmt.Errorf("error reading settings: %w", err)
	}

	timeInterval, err := maputil.GetStringOptional(jsonData, "timeInterval")
	if err != nil {
		return nil, err
	}

	p := client.NewProvider(settings, jsonData, httpClientProvider, cfg, features, plog)
	if err != nil {
		return nil, err
	}

	return &Resource{
		intervalCalculator: intervalv2.NewCalculator(),
		tracer:             tracer,
		log:                plog,
		provider:           p,
		TimeInterval:       timeInterval,
		ID:                 settings.ID,
		URL:                settings.URL,
	}, nil
}

func (r *Resource) Execute(ctx context.Context, headers map[string]string, req *backend.CallResourceRequest) (int, []byte, error) {
	client, err := r.provider.GetClient(headers)
	if err != nil {
		return 500, nil, err
	}

	return r.fetch(ctx, client, req)
}

func (r *Resource) fetch(ctx context.Context, client *client.Client, req *backend.CallResourceRequest) (int, []byte, error) {
	r.log.Debug("Sending resource query", "URL", r.URL)
	u, err := url.Parse(req.URL)
	if err != nil {
		return 500, nil, err
	}

	qs := u.Query()
	if req.Method == http.MethodPost {
		var opts map[string]interface{}
		err = json.Unmarshal(req.Body, &opts)
		if err != nil {
			return 500, nil, err
		}
		for k, v := range opts {
			switch val := v.(type) {
			case string:
				qs.Set(k, val)
			}
		}
	}

	resp, err := client.QueryResource(ctx, req.Method, u.Path, qs)
	if err != nil {
		return 500, nil, err
	}

	defer resp.Body.Close() //nolint

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 500, nil, err
	}
	if resp.StatusCode != 200 {
		return resp.StatusCode, nil, errors.New(string(data))
	}

	return 200, data, err
}
