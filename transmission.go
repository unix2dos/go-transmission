package transmission

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

var (
	ErrNoTorrent = errors.New("No torrent with that id")
)

//TransmissionClient to talk to transmission
type TransmissionClient struct {
	apiclient *ApiClient
}

type Command struct {
	Method    string    `json:"method,omitempty"`
	Arguments arguments `json:"arguments,omitempty"`
	Result    string    `json:"result,omitempty"`
}

type arguments struct {
	Fields           []string     `json:"fields,omitempty"`
	Torrents         Torrents     `json:"torrents,omitempty"`
	Ids              []string     `json:"ids,omitempty"`
	DeleteData       bool         `json:"delete-local-data,omitempty"`
	DownloadDir      string       `json:"download-dir,omitempty"`
	MetaInfo         string       `json:"metainfo,omitempty"`
	Filename         string       `json:"filename,omitempty"`
	TorrentAdded     TorrentAdded `json:"torrent-added"`
	TorrentDuplicate TorrentAdded `json:"torrent-duplicate"`

	// Stats
	ActiveTorrentCount int             `json:"activeTorrentCount"`
	CumulativeStats    cumulativeStats `json:"cumulative-stats"`
	CurrentStats       currentStats    `json:"current-stats"`
	DownloadSpeed      uint64          `json:"downloadSpeed"`
	PausedTorrentCount int             `json:"pausedTorrentCount"`
	TorrentCount       int             `json:"torrentCount"`
	UploadSpeed        uint64          `json:"uploadSpeed"`
	Version            string          `json:"version"`
}

type peer struct {
	Address            string  `json:"address"`
	Name               string  `json:"clientName"`
	Port               int     `json:"port"`
	RateToPeer         int     `json:"rateToPeer"`
	RateToClient       int     `json:"rateToClient"`
	Progress           float32 `json:"progress"`
	Flags              string  `json:"flagStr"`
	IsEncrypted        bool    `json:"isEncrypted"`
	IsUTP              bool    `json:"isUTP"`
	IsUploadingTo      bool    `json:"isUploadingTo"`
	IsIncoming         bool    `json:"isIncoming"`
	IsDownloadingFrom  bool    `json:"isDownloadingFrom"`
	PeerIsInterested   bool    `json:"peerIsInterested"`
	PeerIsChoked       bool    `json:"peerIsChoked"`
	ClientIsInterested bool    `json:"clientIsInterested"`
	ClientIsChoked     bool    `json:"clientIsChoked"`
}

type peers []peer

type tracker struct {
	Announce string `json:"announce"`
	Id       int    `json:"id"`
	Scrape   string `json:"scrape"`
	Tire     int    `json:"tire"`
}

type trackerStat struct {
	Announce              string `json:"announce"`
	AnnounceState         int    `json:"announceState"`
	DownloadCount         int    `json:"downloadCount"`
	HasAnnounced          bool   `json:"hasAnnounced"`
	HasScraped            bool   `json:"hasScraped"`
	Host                  string `json:"host"`
	ID                    uint64 `json:"id"`
	IsBackup              bool   `json:"isBackup"`
	LastAnnouncePeerCount int    `json:"lastAnnouncePeerCount"`
	LastAnnounceResult    string `json:"lastAnnounceResult"`
	LastAnnounceStartTime int64  `json:"lastAnnounceStartTime"`
	LastAnnounceSucceeded bool   `json:"lastAnnounceSucceeded"`
	LastAnnounceTime      int64  `json:"lastAnnounceTime"`
	LastAnnounceTimedOut  bool   `json:"lastAnnounceTimedOut"`
	LastScrapeResult      string `json:"lastScrapeResult"`
	LastScrapeStartTime   int64  `json:"lastScrapeStartTime"`
	LastScrapeSucceeded   bool   `json:"lastScrapeSucceeded"`
	LastScrapeTime        int64  `json:"lastScrapeTime"`
	LastScrapeTimedOut    int64  `json:"lastScrapeTimedOut"`
	LeecherCount          int    `json:"leecherCount"`
	NextAnnounceTime      int64  `json:"nextAnnounceTime"`
	NextScrapeTime        int64  `json:"nextScrapeTime"`
	Scrape                string `json:"scrape"`
	ScrapeState           int    `json:"scrapeState"`
	SeederCount           int    `json:"seederCount"`
	Tier                  int    `json:"tier"`
}

type trackers []tracker

//TorrentAdded data returning
type TorrentAdded struct {
	HashString string `json:"hashString"`
	ID         int    `json:"id"`
	Name       string `json:"name"`
}

// session-stats
type Stats struct {
	ActiveTorrentCount int             `json:"activeCount"`
	CumulativeStats    cumulativeStats `json:"total"`
	CurrentStats       currentStats    `json:"current"`
	DownloadSpeed      uint64          `json:"downloadSpeed"`
	PausedTorrentCount int             `json:"pausedCount"`
	TorrentCount       int             `json:"count"`
	UploadSpeed        uint64          `json:"uploadSpeed"`
}
type cumulativeStats struct {
	DownloadedBytes uint64        `json:"downloadedBytes"`
	FilesAdded      int           `json:"filesAdded"`
	SecondsActive   time.Duration `json:"secondsActive"`
	SessionCount    int           `json:"sessionCount"`
	UploadedBytes   uint64        `json:"uploadedBytes"`
}
type currentStats struct {
	DownloadedBytes uint64        `json:"downloadedBytes"`
	FilesAdded      int           `json:"filesAdded"`
	SecondsActive   time.Duration `json:"secondsActive"`
	SessionCount    int           `json:"sessionCount"`
	UploadedBytes   uint64        `json:"uploadedBytes"`
}

func (s *Stats) CurrentActiveTime() string {
	return (time.Second * s.CurrentStats.SecondsActive).String()
}

func (s *Stats) CumulativeActiveTime() string {
	return (time.Second * s.CumulativeStats.SecondsActive).String()
}

type File struct {
	Completed int64  `json:"bytesCompleted"`
	Size      int64  `json:"length"`
	Name      string `json:"name"`
}

type Files []File

//Torrent struct for torrents
type Torrent struct {
	ID              int           `json:"id"`
	Name            string        `json:"name"`
	Status          Status        `json:"status"`
	AddedDate       int64         `json:"addedDate"` // unix timestamp
	StartDate       int64         `json:"startDate"` // unix timestamp
	DoneDate        int64         `json:"doneDate"`  // unix timestamp
	LeftUntilDone   uint64        `json:"leftUntilDone"`
	SizeWhenDone    uint64        `json:"sizeWhenDone"`
	Eta             time.Duration `json:"eta"` // in seconds, not a valid time.Duration
	UploadRatio     float64       `json:"uploadRatio"`
	RateDownload    uint64        `json:"rateDownload"`
	RateUpload      uint64        `json:"rateUpload"`
	DownloadDir     string        `json:"downloadDir"`
	DownloadedEver  uint64        `json:"downloadedEver"`
	UploadedEver    uint64        `json:"uploadedEver"`
	HaveUnchecked   uint64        `json:"haveUnchecked"`
	HaveValid       uint64        `json:"haveValid"`
	IsFinished      bool          `json:"isFinished"`
	PercentDone     float32       `json:"percentDone"` // 0...1, double
	SeedRatioMode   int           `json:"seedRatioMode"`
	Files           Files         `json:"files"`
	Peers           peers         `json:"peers"`
	Trackers        trackers      `json:"trackers"`
	TrackerStats    []trackerStat `json:"trackerStats"`
	Error           int           `json:"error"`
	ErrorString     string        `json:"errorString"`
	InfoHash        string        `json:"hashString"`
	TotalSize       uint64        `json:"totalSize"`
	DownloadSeconds uint64        `json:"secondsDownloading"`
	SeedSeconds     uint64        `json:"secondsSeeding"`
}

func (t *Torrent) GetSize() uint64 {
	return t.TotalSize
}

func (t *Torrent) GetPercent() float32 {
	return t.PercentDone * 100
}

// Ratio returns the upload ratio of the torrent
func (t *Torrent) Ratio() string {
	if t.UploadRatio < 0 {
		return "∞"
	}
	return fmt.Sprintf("%.3f", t.UploadRatio)
}

// ETA returns the time left for the download to finish
func (t *Torrent) ETA() string {
	if t.Eta < 0 {
		return "∞"
	}
	return (time.Second * t.Eta).String()
}

// GetTrackers combines the torrent's trackers in one string
func (t *Torrent) GetTrackers() string {
	buf := new(bytes.Buffer)
	for i := range t.Trackers {
		buf.WriteString(fmt.Sprintf("%s\n", t.Trackers[i].Announce))
	}

	return buf.String()
}

// Have returns haveValid + haveUnchecked
func (t *Torrent) Have() uint64 {
	return t.HaveValid + t.HaveUnchecked
}

func (t *Torrent) IsCompleted() bool {
	return t.PercentDone == 1
}

// Torrents represent []Torrent
type Torrents []*Torrent

// GetIDs returns []int of all the ids
func (t Torrents) GetIDs() []string {
	ids := make([]string, 0, len(t))
	for i := range t {
		ids = append(ids, t[i].InfoHash)
	}
	return ids
}

// sortType keeps track of which sorting we are using
var sortType = SortID // SortID is transmission's default

// SetSort takes a 'Sorting' to set 'sortType'
func (ac *TransmissionClient) SetSort(st Sorting) {
	sortType = st
}

//New create new transmission torrent
func New(url string, username string, password string) (*TransmissionClient, error) {
	apiclient := NewClient(url, username, password)
	client := &TransmissionClient{apiclient: apiclient}

	// test that we have a working client
	cmd := Command{Method: "session-get"}
	_, err := client.sendCommand(cmd)
	if err != nil {
		return client, err
	}

	return client, nil

}

//GetTorrents get a list of torrents
func (ac *TransmissionClient) GetTorrents() (Torrents, error) {
	cmd := NewGetTorrentsCmd()

	out, err := ac.ExecuteCommand(cmd)
	if err != nil {
		return nil, err
	}

	torrents := out.Arguments.Torrents

	// sorting
	switch sortType {
	case SortID:
		return torrents, nil // already sorted by ID
	case SortRevID:
		torrents.SortID(true)
	case SortName:
		torrents.SortName(false)
	case SortRevName:
		torrents.SortName(true)
	case SortAge:
		torrents.SortAge(false)
	case SortRevAge:
		torrents.SortAge(true)
	case SortSize:
		torrents.SortSize(false)
	case SortRevSize:
		torrents.SortSize(true)
	case SortProgress:
		torrents.SortProgress(false)
	case SortRevProgress:
		torrents.SortProgress(true)
	case SortDownSpeed:
		torrents.SortDownSpeed(false)
	case SortRevDownSpeed:
		torrents.SortDownSpeed(true)
	case SortUpSpeed:
		torrents.SortUpSpeed(false)
	case SortRevUpSpeed:
		torrents.SortUpSpeed(true)
	case SortDownloaded:
		torrents.SortDownloaded(false)
	case SortRevDownloaded:
		torrents.SortDownloaded(true)
	case SortUploaded:
		torrents.SortUploaded(false)
	case SortRevUploaded:
		torrents.SortUploaded(true)
	case SortRatio:
		torrents.SortRatio(false)
	case SortRevRatio:
		torrents.SortRatio(true)
	}

	return torrents, nil
}

// GetTorrent takes an id and returns *Torrent
func (ac *TransmissionClient) GetTorrent(id string) (*Torrent, error) {
	cmd := NewGetTorrentsCmd()
	cmd.Arguments.Ids = append(cmd.Arguments.Ids, id)

	out, err := ac.ExecuteCommand(cmd)
	if err != nil {
		return &Torrent{}, err
	}

	if len(out.Arguments.Torrents) > 0 {
		return out.Arguments.Torrents[0], nil
	}
	return &Torrent{}, ErrNoTorrent
}

// Delete takes a bool, if true it will delete with data;
// returns the name of the deleted torrent if it succeed
func (ac *TransmissionClient) DeleteTorrent(id string, withData bool) (string, error) {
	torrent, err := ac.GetTorrent(id)
	if err != nil {
		return "", err
	}

	cmd := newDelCmd(id, withData)

	_, err = ac.ExecuteCommand(cmd)
	if err != nil {
		return "", err
	}

	return torrent.Name, nil
}

// GetStats returns "session-stats"
func (ac *TransmissionClient) GetStats() (*Stats, error) {
	cmd := &Command{
		Method: "session-stats",
	}

	out, err := ac.ExecuteCommand(cmd)
	if err != nil {
		return nil, err
	}

	return &Stats{
		ActiveTorrentCount: out.Arguments.ActiveTorrentCount,
		CumulativeStats:    out.Arguments.CumulativeStats,
		CurrentStats:       out.Arguments.CurrentStats,
		DownloadSpeed:      out.Arguments.DownloadSpeed,
		PausedTorrentCount: out.Arguments.PausedTorrentCount,
		TorrentCount:       out.Arguments.TorrentCount,
		UploadSpeed:        out.Arguments.UploadSpeed,
	}, nil
}

//StartTorrent start the torrent
func (ac *TransmissionClient) StartTorrent(ids ...string) (string, error) {
	return ac.sendSimpleCommand("torrent-start", ids...)
}

//StopTorrent start the torrent
func (ac *TransmissionClient) StopTorrent(ids ...string) (string, error) {
	return ac.sendSimpleCommand("torrent-stop", ids...)
}

// VerifyTorrent verifies a torrent
func (ac *TransmissionClient) VerifyTorrent(ids ...string) (string, error) {
	return ac.sendSimpleCommand("torrent-verify", ids...)
}

// StartAll starts all the torrents
func (ac *TransmissionClient) StartAll() error {
	cmd := Command{Method: "torrent-start"}
	torrents, err := ac.GetTorrents()
	if err != nil {
		return err
	}

	cmd.Arguments.Ids = torrents.GetIDs()
	if _, err := ac.sendCommand(cmd); err != nil {
		return err
	}

	return nil
}

// StopAll stops all torrents
func (ac *TransmissionClient) StopAll() error {
	cmd := Command{Method: "torrent-stop"}
	torrents, err := ac.GetTorrents()
	if err != nil {
		return err
	}

	cmd.Arguments.Ids = torrents.GetIDs()
	if _, err := ac.sendCommand(cmd); err != nil {
		return err
	}

	return nil
}

// VerifyAll verfies all torrents
func (ac *TransmissionClient) VerifyAll() error {
	cmd := Command{Method: "torrent-verify"}

	torrents, err := ac.GetTorrents()
	if err != nil {
		return err
	}

	cmd.Arguments.Ids = torrents.GetIDs()
	if _, err := ac.sendCommand(cmd); err != nil {
		return err
	}

	return nil
}

func NewGetTorrentsCmd() *Command {
	cmd := &Command{}

	cmd.Method = "torrent-get"
	cmd.Arguments.Fields = []string{"id", "name", "hashString", "status", "addedDate", "startDate", "doneDate",
		"leftUntilDone", "sizeWhenDone", "haveValid", "haveUnchecked", "isFinished", "percentDone", "eta",
		"rateDownload", "rateUpload", "downloadDir", "downloadedEver", "uploadRatio", "uploadedEver",
		"seedRatioMode", "error", "errorString", "files", "peers", "trackers", "trackerStats", "totalSize",
		"secondsDownloading", "secondsSeeding"}
	// cmd.Arguments.Fields = []string{
	// 	"activityDate",
	// 	"addedDate",
	// 	"bandwidthPriority",
	// 	"comment",
	// 	"corruptEver",
	// 	"creator",
	// 	"dateCreated",
	// 	"desiredAvailable",
	// 	"doneDate",
	// 	"downloadDir",
	// 	"downloadedEver",
	// 	"downloadLimit",
	// 	"downloadLimited",
	// 	"error",
	// 	"errorString",
	// 	"eta",
	// 	"etaIdle",
	// 	"files",
	// 	"fileStats",
	// 	"hashString",
	// 	"haveUnchecked",
	// 	"haveValid",
	// 	"honorsSessionLimits",
	// 	"id",
	// 	"isFinished",
	// 	"isPrivate",
	// 	"isStalled",
	// 	"leftUntilDone",
	// 	"magnetLink",
	// 	"manualAnnounceTime",
	// 	"maxConnectedPeers",
	// 	"metadataPercentComplete",
	// 	"name",
	// 	"peer-limit",
	// 	"peers",
	// 	"peersConnected",
	// 	"peersFrom",
	// 	"peersGettingFromUs",
	// 	"peersSendingToUs",
	// 	"percentDone",
	// 	"pieces",
	// 	"pieceCount",
	// 	"pieceSize",
	// 	"priorities",
	// 	"queuePosition",
	// 	"rateDownload",
	// 	"rateUpload",
	// 	"recheckProgress",
	// 	"secondsDownloading",
	// 	"secondsSeeding",
	// 	"seedIdleLimit",
	// 	"seedIdleMode",
	// 	"seedRatioLimit",
	// 	"seedRatioMode",
	// 	"sizeWhenDone",
	// 	"startDate",
	// 	"status",
	// 	"trackers",
	// 	"trackerStats",
	// 	"totalSize",
	// 	"torrentFile",
	// 	"uploadedEver",
	// 	"uploadLimit",
	// 	"uploadLimited",
	// 	"uploadRatio",
	// 	"wanted",
	// 	"webseeds",
	// 	"webseedsSendingToUs",
	// }

	return cmd
}

func NewAddCmd() *Command {
	cmd := &Command{}
	cmd.Method = "torrent-add"
	return cmd
}

// URL or magnet
func NewAddCmdByURL(url string) *Command {
	cmd := NewAddCmd()
	cmd.Arguments.Filename = url
	return cmd
}

func NewAddCmdByFilename(filename string) *Command {
	cmd := NewAddCmd()
	cmd.Arguments.Filename = filename
	return cmd
}

func NewAddCmdByFile(file string) (*Command, error) {
	cmd := NewAddCmd()

	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cmd.Arguments.MetaInfo = base64.StdEncoding.EncodeToString(fileData)

	return cmd, nil
}

func NewAddCmdByBytes(b []byte) (*Command, error) {
	cmd := NewAddCmd()

	cmd.Arguments.MetaInfo = base64.StdEncoding.EncodeToString(b)

	return cmd, nil
}

func (cmd *Command) SetDownloadDir(dir string) {
	cmd.Arguments.DownloadDir = dir
}

func newDelCmd(id string, removeFile bool) *Command {
	cmd := &Command{}
	cmd.Method = "torrent-remove"
	cmd.Arguments.Ids = []string{id}
	cmd.Arguments.DeleteData = removeFile
	return cmd
}

func (ac *TransmissionClient) ExecuteCommand(cmd *Command) (*Command, error) {
	out := &Command{}

	body, err := json.Marshal(cmd)
	if err != nil {
		return out, err
	}
	output, err := ac.apiclient.Post(string(body))
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(output, &out)
	if err != nil {
		log.Printf("output: %s", output)
		return out, err
	}

	return out, nil
}

func (ac *TransmissionClient) ExecuteAddCommand(addCmd *Command) (TorrentAdded, error) {
	outCmd, err := ac.ExecuteCommand(addCmd)
	if err != nil {
		return TorrentAdded{}, err
	}
	if outCmd.Arguments.TorrentDuplicate.HashString != "" {
		return outCmd.Arguments.TorrentDuplicate, nil
	}
	return outCmd.Arguments.TorrentAdded, nil
}

func encodeFile(file string) (string, error) {
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(fileData), nil
}

// Version returns transmission's version
func (ac *TransmissionClient) Version() string {
	cmd := Command{Method: "session-get"}

	resp, _ := ac.sendCommand(cmd)
	return resp.Arguments.Version
}

func (ac *TransmissionClient) sendSimpleCommand(method string, ids ...string) (result string, err error) {
	cmd := Command{Method: method}
	cmd.Arguments.Ids = append([]string{}, ids...)
	resp, err := ac.sendCommand(cmd)
	return resp.Result, err
}

func (ac *TransmissionClient) sendCommand(cmd Command) (response Command, err error) {
	var body, output []byte
	body, err = json.Marshal(cmd)
	if err != nil {
		return
	}
	output, err = ac.apiclient.Post(string(body))
	if err != nil {
		return
	}
	// l.Infof("output %s", output)
	err = json.Unmarshal(output, &response)
	if err != nil {
		return
	}
	return response, nil
}
