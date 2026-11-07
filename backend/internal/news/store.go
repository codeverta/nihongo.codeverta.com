package news

import "context"

// ImageStore keeps article images in durable object storage so reads never
// depend on the publisher's server. Put receives an already optimized JPEG and
// returns its public URL. Owns reports whether a stored imageUrl is managed by
// this store.
type ImageStore interface {
	Put(ctx context.Context, jpeg []byte) (string, error)
	Owns(imageURL string) bool
}
