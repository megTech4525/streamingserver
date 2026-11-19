A small Go service that adds accounts, stream keys, and publish authorization
to a self-hosted MediaMTX live streaming setup.

OBS → MediaMTX → (asks this API: "may this key publish?") → viewers watch via HLS

Status: early. Works end to end locally
