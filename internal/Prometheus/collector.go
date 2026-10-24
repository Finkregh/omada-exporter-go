package Prometheus

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"omada_exporter_go/internal/Omada/Model/AccessPoint"
	"omada_exporter_go/internal/Omada/Model/Devices"
	"omada_exporter_go/internal/Omada/Model/Gateway"
	"omada_exporter_go/internal/Omada/Model/Interface"
	"omada_exporter_go/internal/Omada/Model/Switch"
)

var (
	omada_scrape_duration = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "omada_scrape_duration",
			Help: "The amount of time it took to scrape the Omada controller in seconds",
		},
	)
)

// CollectMetrics scrapes the controller and exposes whatever it could collect.
// A failing device class does not abort the scrape (the other classes still
// report), but the failures are joined into the returned error so the caller can
// log them instead of silently serving a half-empty metric set.
func CollectMetrics() error {
	start := time.Now()
	deviceList, err := Devices.Get()
	if err != nil {
		return err
	}

	var (
		omadaDevices []Interface.Device
		errs         []error
	)

	switches, err := Switch.Get(*deviceList)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get switches: %w", err))
	} else {
		for i := range *switches {
			(*switches)[i].AppendGeneralProperties(deviceList)
		}
		if err := Interface.AppendDevicesSlice(&omadaDevices, *switches); err != nil {
			errs = append(errs, err)
		}
	}

	gateways, err := Gateway.Get(*deviceList)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get gateways: %w", err))
	} else {
		for i := range *gateways {
			(*gateways)[i].AppendGeneralProperties(deviceList)
		}
		if err := Interface.AppendDevicesSlice(&omadaDevices, *gateways); err != nil {
			errs = append(errs, err)
		}
	}

	aps, err := AccessPoint.Get(*deviceList)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get access points: %w", err))
	} else {
		for i := range *aps {
			(*aps)[i].AppendGeneralProperties(deviceList)
		}
		if err := Interface.AppendDevicesSlice(&omadaDevices, *aps); err != nil {
			errs = append(errs, err)
		}
	}

	omada_scrape_duration.Set(time.Since(start).Seconds())

	ExposeDeviceMetrics(omadaDevices)
	ExposePortMetrics(omadaDevices)
	ExposeRadioMetrics(omadaDevices)

	return errors.Join(errs...)
}
