# memprofiler
[![Build Status](https://travis-ci.org/memprofiler/memprofiler.svg?branch=master)](https://travis-ci.org/memprofiler/memprofiler)
[![codecov](https://codecov.io/gh/memprofiler/memprofiler/branch/master/graph/badge.svg)](https://codecov.io/gh/memprofiler/memprofiler)

Memprofiler helps to track memory allocations of your Go applications on 
large time intervals. Go runtime implements multiple memory management 
optimizations in order to achieve good performance, low heap allocation 
cost and high degree of memory reuse. Therefore, sometimes it may be 
tricky to distinguish "normal" runtime behaviour from real memory leak.
If you have doubts whether your Go service is leaking, you're on the right
track. Memprofiler aims to be an open source equivalent of 
[stackimpact.com](https://stackimpact.com/). 

<aside class="warning">
The project is under active development and not ready for usage yet.
<aside>


## Getting started

Memprofiler is a client-server application. Memprofiler client is embedded 
into your Go service and streams memory usage reports to the Memprofiler server. 
Memprofiler server stores reports and performs some computations on the 
data stream to turn it in a small set of aggregated metrics. 
User will be able to interact with Memprofiler server via simple Web UI.

![Components](https://cdn1.imggmi.com/uploads/2019/4/4/0ef08a39ffbe10ca7279b04c6eedc4b0-full.png)

To use memprofiler in your application, run client in your `main` function:

```go
package example

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/client"
	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/utils"
)

func main() {
	// prepare client configuration
	cfg := &client.Config{
		// server address
		ServerEndpoint: "localhost:46219",
		// description of your application instance
		ServiceDescription: &schema.ServiceDescription{
			ServiceType:     "test_application",
			ServiceInstance: "node_1",
		},
		// granularity
		Periodicity: &utils.Duration{Duration: time.Second},
		// logging setting
		Verbose: false,
	}

	// you can implement your own logger
	log := client.LoggerFromLogrus(logrus.New())

	// run profiler and stop it explicitly on exit
	profiler, err := client.NewProfiler(log, cfg)
	if err != nil {
		panic(err)
	}
	defer profiler.Quit()

	// ...
}

```

To run Memprofiler server, just install it and prepare server config 
(you can refer to [config example](https://github.com/memprofiler/memprofiler/blob/master/server/config/example.yml)).

```bash
 ✗ GO111MODULE=on go get github.com/memprofiler/memprofiler
 ✗ memprofiler -c config.yml 
DEBU[0000] Starting storage                             
DEBU[0000] Starting metrics computer                    
INFO[0000] HTTP Frontend server resource                 URL=/schema.MemprofilerFrontend/GetSessions subsystem=frontend
INFO[0000] HTTP Frontend server resource                 URL=/schema.MemprofilerFrontend/GetServices subsystem=frontend
INFO[0000] HTTP Frontend server resource                 URL=/schema.MemprofilerFrontend/GetInstances subsystem=frontend
INFO[0000] HTTP Frontend server resource                 URL=/schema.MemprofilerFrontend/SubscribeForSession subsystem=frontend
INFO[0000] Starting service                              service=backend
INFO[0000] Starting service                              service=frontend

```
