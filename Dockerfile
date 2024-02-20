# Copyright 2021 Richard Kosegi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.21 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY config.go config.go
COPY client.go client.go
COPY types.go types.go
COPY exporter.go exporter.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o owm-exporter . ; strip owm-exporter

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/owm-exporter /
USER 65532:65532

EXPOSE 9111

ENTRYPOINT ["/owm-exporter"]
