package adapter

import (
	"log"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"fritzcable_exporter/internal/client"
)

type BoxCollector struct {
	Client *client.Client
}

var (
	readyStateDesc = prometheus.NewDesc(
		"docsis_ready_state",
		"ready state of docsis connection",
		[]string{"state"}, nil,
	)
	correctableErrorsDesc = prometheus.NewDesc(
		"docsis_channel_correctable_errors_total",
		"number of correctable errors on channel",
		[]string{"channelType", "channelID", "frequency", "modulation"}, nil,
	)
	uncorrectableErrorsDesc = prometheus.NewDesc(
		"docsis_channel_uncorrectable_errors_total",
		"number of uncorrectable errors on channel",
		[]string{"chanelType", "channelID", "frequency", "modulation"}, nil,
	)
)

func (bc BoxCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(bc, ch)
}

func (bc BoxCollector) Collect(ch chan<- prometheus.Metric) {
	ok, err := bc.Client.CheckLogin()
	if !ok || err != nil {
		if inErr := bc.Client.Login(); inErr != nil {
			log.Print("failed to login to collector metrics", inErr)
			return
		}
	}
	data, err := bc.Client.DocsisStats()
	if err != nil {
		log.Println("could not get upstream data: ", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		readyStateDesc,
		prometheus.GaugeValue,
		float64(1),
		data.ReadyState,
	)
	for channelType, channels := range data.ChannelDs {
		for _, channel := range channels {
			ch <- prometheus.MustNewConstMetric(
				correctableErrorsDesc,
				prometheus.CounterValue,
				float64(channel.CorrectableErrors),
				channelType, strconv.Itoa(channel.ChannelID), channel.Frequency, channel.Modulation,
			)
			ch <- prometheus.MustNewConstMetric(
				uncorrectableErrorsDesc,
				prometheus.CounterValue,
				float64(channel.NonCorrectableErrors),
				channelType, strconv.Itoa(channel.ChannelID), channel.Frequency, channel.Modulation,
			)
		}
	}
}
