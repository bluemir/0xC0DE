package store

type Metadata struct {
	Kind      string            `json:"kind"`
	Id        string            `json:"id"`
	Namespace string            `json:"namespace"`
	Rev       int               `json:"rev"`
	Labels    map[string]string `json:"labels"`
}

func (meta *Metadata) GetMetadata() *Metadata {
	return meta
}
