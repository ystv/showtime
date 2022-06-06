package livestream

import "context"

type (
	// RTMPOutput is a simple send livestream to an RTMP endpoint.
	RTMPOutput struct {
		ID        int    `db:"rtmp_output_id"`
		OutputURL string `db:"output_url"`
	}
)

// NewRTMPOutput creates a new custom RTMP stream to an endpoint.
func (ls *Livestreamer) NewRTMPOutput(ctx context.Context, outputURL string) (RTMPOutput, error) {
	custom := RTMPOutput{}
	err := ls.db.GetContext(ctx, &custom.ID, `
		INSERT INTO rtmp_outputs (output_url)
		VALUES ($1) RETURNING rtmp_output_id;
	`, outputURL)
	return custom, err
}

// GetRTMPOutput retrives an RTMP output by ID.
func (ls *Livestreamer) GetRTMPOutput(ctx context.Context, rtmpOutputID int) (RTMPOutput, error) {
	custom := RTMPOutput{}
	err := ls.db.GetContext(ctx, &custom, `
		SELECT rtmp_output_id, output_url
		FROM rtmp_outputs
		WHERE rtmp_output_id = $1;
	`, rtmpOutputID)
	return custom, err
}

// DeleteRTMPOutput deletes a RTMP output by ID.
func (ls *Livestreamer) DeleteRTMPOutput(ctx context.Context, rtmpOutputID int) error {
	_, err := ls.db.ExecContext(ctx, `
		DELETE FROM rtmp_outputs
		WHERE rtmp_output_id = $1;
	`, rtmpOutputID)
	return err
}
