package main

type Response struct {
	Body string
}

func (resp *Response) toBytes() []byte {
	return []byte(resp.Body)
}

var channels = make(map[string](chan Response))

type ChannelRegistry struct {
}

func (reg *ChannelRegistry) Ensure(key string) chan Response {
	channel := reg.Get(key)
	if channel == nil {
		reg.Create(key)
		channel = reg.Get(key)
	}
	return channel
}

func (reg *ChannelRegistry) Get(key string) chan Response {
	return channels[key]
}

func (reg *ChannelRegistry) Create(key string) {
	channels[key] = make(chan Response, 1)
}
