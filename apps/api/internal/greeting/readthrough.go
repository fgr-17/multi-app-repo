package greeting

// ReadThrough serves GET from the Mongo read model and falls back to
// replaying the Postgres event stream when the projection is behind or empty.
type ReadThrough struct {
	Writes Store
	Reads  *MongoView
}

func (r ReadThrough) Get() (Record, error) {
	if r.Reads != nil {
		rec, err := r.Reads.Get()
		if err == nil && rec.Name != "" {
			return rec, nil
		}
	}
	return r.Writes.Get()
}

func (r ReadThrough) SaveIfNewer(incoming Record) (Record, bool, error) {
	return r.Writes.SaveIfNewer(incoming)
}

func (r ReadThrough) Ping() error {
	return r.Writes.Ping()
}
