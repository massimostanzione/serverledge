package lb

import "net/url"

const MULT_FACTOR = 10

// Rappresenta un server con un peso associato
type Server struct {
	target *url.URL
	weight int
}
