// Copyright 2021 Richard Kosegi
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"encoding/json"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	baseUri = "https://api.openweathermap.org/data/2.5/weather"
)

type CacheEntry struct {
	lastResponse *ApiResponse
	lastUpdate   time.Time
}

type ApiResponse struct {
	Main struct {
		Temp     float32 `json:"temp"`
		TempMin  float32 `json:"temp_min"`
		TempMax  float32 `json:"temp_max"`
		TempFeel float32 `json:"feels_like"`
		Pressure int     `json:"pressure"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed     float32 `json:"speed"`
		Direction int     `json:"deg"`
	} `json:"wind"`
	Name string `json:"name"`
}

type OwmClient struct {
	AppId   string
	metrics ExporterMetrics
}

func NewClient(apiKey string, metrics ExporterMetrics) OwmClient {
	return OwmClient{
		AppId:   apiKey,
		metrics: metrics,
	}
}

var (
	cache = map[string]CacheEntry{}
)

func (client *OwmClient) Fetch(_ context.Context, target Target, logger log.Logger) (*ApiResponse, error) {
	var fetch = true
	last, present := cache[target.Name]
	if present {
		if time.Now().Unix() < int64(target.Interval)+last.lastUpdate.Unix() {
			fetch = false
		}
	}
	if fetch {

		var uri = fmt.Sprintf("%s?lat=%s&lon=%s&units=metric&appid=%s",
			baseUri, target.Latitude, target.Longitude, client.AppId)

		//nolint:errcheck
		level.Info(logger).Log("msg", "Fetching current conditions", "name", target.Name)

		resp, err := http.Get(uri)

		client.metrics.ApiRequests.WithLabelValues(target.Name).Inc()

		//nolint:errcheck
		level.Debug(logger).Log("msg", "Got response from API", "code", resp.StatusCode)

		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				//nolint:errcheck
				level.Warn(logger).Log("msg", "failed to close body", "error", err)
			}
		}(resp.Body)
		var apiResponse ApiResponse
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			return nil, err
		} else {
			cache[target.Name] = CacheEntry{
				lastResponse: &apiResponse,
				lastUpdate:   time.Now(),
			}
			return &apiResponse, nil
		}

	} else {
		//nolint:errcheck
		level.Debug(logger).Log("msg", "Results are being served from cache", "name", target.Name)
		return last.lastResponse, nil
	}
}
