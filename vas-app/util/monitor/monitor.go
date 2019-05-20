package monitor

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ResponseTime     *prometheus.HistogramVec
	_DurationSecs     *prometheus.HistogramVec
	_VideoCuts        *prometheus.HistogramVec
	_RequestsCounter  *prometheus.CounterVec
	_RequestsParallel *prometheus.GaugeVec
)

func Init(namespace, subsystem string) {
	_ResponseTime = func() *prometheus.HistogramVec {
		vec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "response_time",
				Help:      "Response time of requests",
				Buckets: []float64{
					0, 0.1, 0.2, 0.3, 0.5, 1, 5, 10, 60, 300, 1800, 3600,
				}, // time.second
			},
			[]string{"api", "code"},
		)
		prometheus.MustRegister(vec)
		return vec
	}()

	_DurationSecs = func() *prometheus.HistogramVec {
		vec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "duration_secs",
				Help:      "duration of video",
				Buckets: []float64{
					0, 300, 600, 1200, 1800, 2400, 3000, 3600, 5400, 7200, 18000,
				}, // time.second
			},
			[]string{"api"},
		)
		prometheus.MustRegister(vec)
		return vec
	}()

	_VideoCuts = func() *prometheus.HistogramVec {
		vec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "video_cuts",
				Help:      "cuts of video",
				Buckets: []float64{
					0, 300, 600, 1200, 1800, 2400, 3000, 3600, 5400, 7200, 18000,
				}, // time.second
			},
			[]string{"api"},
		)
		prometheus.MustRegister(vec)
		return vec
	}()

	_RequestsCounter = func() *prometheus.CounterVec {
		vec := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_counter",
				Help:      "number of requests",
			},
			[]string{"api", "code"},
		)
		prometheus.MustRegister(vec)
		return vec
	}()

	_RequestsParallel = func() *prometheus.GaugeVec {
		vec := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_parallel",
				Help:      "parallel of requests",
			},
			[]string{"api"},
		)
		prometheus.MustRegister(vec)
		return vec
	}()
}

func ResponseTime(api string, code int) prometheus.Histogram {
	if _ResponseTime == nil {
		return nil
	}
	return _ResponseTime.WithLabelValues(api, strconv.Itoa(code))
}

func ResponseTimeFrom(api string, code int, start time.Time) {
	respTime := ResponseTime(api, code)
	if respTime == nil {
		return
	}
	respTime.Observe(time.Since(start).Seconds())
}

func DurationSecs(api string) prometheus.Histogram {
	if _DurationSecs == nil {
		return nil
	}
	return _DurationSecs.WithLabelValues(api)
}

func DurationSecsSet(api string, secs float64) {
	durSecs := DurationSecs(api)
	if durSecs == nil {
		return
	}
	durSecs.Observe(secs)
}

func VideoCuts(api string) prometheus.Histogram {
	if _VideoCuts == nil {
		return nil
	}
	return _VideoCuts.WithLabelValues(api)
}

func VideoCutsSet(api string, num int32) {
	videoCuts := VideoCuts(api)
	if videoCuts == nil {
		return
	}
	videoCuts.Observe(float64(num))
}

func RequestsCounter(api string, code int) prometheus.Counter {
	if _RequestsCounter == nil {
		return nil
	}
	return _RequestsCounter.WithLabelValues(api, strconv.Itoa(code))
}

func RequestsCounterInc(api string, code int) {
	reqCounter := RequestsCounter(api, code)
	if reqCounter == nil {
		return
	}
	reqCounter.Inc()
}

func RequestsParallel(api string) prometheus.Gauge {
	if _RequestsParallel == nil {
		return nil
	}
	return _RequestsParallel.WithLabelValues(api)
}

func RequestsParallelInc(api string) {
	requestsParallel := RequestsParallel(api)
	if requestsParallel == nil {
		return
	}
	requestsParallel.Inc()
}

func RequestsParallelDec(api string) {
	requestsParallel := RequestsParallel(api)
	if requestsParallel == nil {
		return
	}
	requestsParallel.Dec()
}
