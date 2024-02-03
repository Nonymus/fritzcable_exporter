package client

// The interesting parts /data.lua replies for Docsis stats

type Container struct {
	Data Data `json:"data"`
}

type Data struct {
	ChannelDs  map[string][]DownChannelData `json:"channelDs"`
	ChannelUs  map[string][]UpChannelData   `json:"channelUs"`
	OEM        string                       `json:"OEM"`
	ReadyState string                       `json:"readyState"`
}

type CommonChannelData struct {
	ChannelID  int    `json:"channelID"`
	Frequency  string `json:"frequency"`
	Modulation string `json:"modulation"`
	PowerLevel string `json:"powerLevel"`
}

type DownChannelData struct {
	CommonChannelData
	CorrectableErrors    int     `json:"corrErrors"`
	NonCorrectableErrors int     `json:"nonCorrErrors"`
	Latency              float64 `json:"latency"`
	MSE                  string  `json:"mse"`
}

type UpChannelData struct {
	CommonChannelData
	Multiplex string `json:"multiplex"`
}
