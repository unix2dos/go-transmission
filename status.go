package transmission

type Status int

const (
	TrStopped Status = iota
	TrCheckPending
	TrChecking
	TrDownloadPending
	TrDownloading
	TrSeedPending
	TrSeeding
)

// Status translates the status of the torrent
func (ts Status) String() string {
	switch ts {
	case TrStopped:
		return "Stopped"
	case TrCheckPending:
		return "Check waiting"
	case TrChecking:
		return "Checking"
	case TrDownloadPending:
		return "Download waiting"
	case TrDownloading:
		return "Downloading"
	case TrSeedPending:
		return "Seed waiting"
	case TrSeeding:
		return "Seeding"
	default:
		return "unknown"
	}
}

func (ts Status) IsStarted() bool {
	return ts == TrChecking ||
		ts == TrDownloadPending ||
		ts == TrDownloading ||
		ts == TrSeedPending ||
		ts == TrSeeding
}
