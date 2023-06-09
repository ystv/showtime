package mixer

import "github.com/jmoiron/sqlx"

type (
	Mixerer struct {
		db *sqlx.DB
	}
	// EditLivestream are parameters required to create or update a livestream.
	EditOBSMixer struct {
		Address  string `db:"address"`
		Username string `db:"username"`
		Password string `db:"password"`
	}
	// Livestream is the metadata of a stream and the links to external
	// platforms.
	OBSMixer struct {
		ID       int    `db:"mixer_id" json:"livestreamID"`
		Address  string `db:"address"`
		Username string `db:"username"`
		Password string `db:"password"`
	}
)
