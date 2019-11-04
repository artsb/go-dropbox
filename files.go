package dropbox

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Files client for files and folders.
type Files struct {
	*Client
}

// NewFiles client.
func NewFiles(config *Config) *Files {
	return &Files{
		Client: &Client{
			Config: config,
		},
	}
}

// WriteMode determines what to do if the file already exists.
type WriteMode string

// Supported write modes.
const (
	WriteModeAdd       WriteMode = "add"
	WriteModeOverwrite WriteMode = "overwrite"
	WriteModeUpdate    WriteMode = "update"
)

// writeModeUpdate specifies a write mode update.
type writeModeUpdate struct {
	Tag string `json:".tag"`
	Rev string `json:"update"`
}

// newWriteModeUpdate creates writeModeUpdate with given rev.
func newWriteModeUpdate(rev string) *writeModeUpdate {
	return &writeModeUpdate{
		Tag: "update",
		Rev: rev,
	}
}

// Dimensions specifies the dimensions of a photo or video.
type Dimensions struct {
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
}

// GPSCoordinates specifies the GPS coordinate of a photo or video.
type GPSCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// PhotoMetadata specifies metadata for a photo.
type PhotoMetadata struct {
	Dimensions *Dimensions     `json:"dimensions,omitempty"`
	Location   *GPSCoordinates `json:"location,omitempty"`
	TimeTaken  time.Time       `json:"time_taken,omitempty"`
}

// VideoMetadata specifies metadata for a video.
type VideoMetadata struct {
	Dimensions *Dimensions     `json:"dimensions,omitempty"`
	Location   *GPSCoordinates `json:"location,omitempty"`
	TimeTaken  time.Time       `json:"time_taken,omitempty"`
	Duration   uint64          `json:"duration,omitempty"`
}

// MediaMetadata provides metadata for a photo or video.
type MediaMetadata struct {
	Photo *PhotoMetadata `json:"photo,omitempty"`
	Video *VideoMetadata `json:"video,omitempty"`
}

// MediaInfo provides additional information for a photo or video file.
type MediaInfo struct {
	Pending  bool           `json:"pending"`
	Metadata *MediaMetadata `json:"metadata,omitempty"`
}

// FileSharingInfo for a file which is contained in a shared folder.
type FileSharingInfo struct {
	ReadOnly             bool   `json:"read_only"`
	ParentSharedFolderID string `json:"parent_shared_folder_id"`
	ModifiedBy           string `json:"modified_by,omitempty"`
}

// PropertyField specifies additional property field.
type PropertyField struct {
	Name  string `json:"name"`  // Max length 256 bytes.
	Value string `json:"value"` // Max length 1024 bytes.
}

// PropertyGroup specifies additional property group.
type PropertyGroup struct {
	TemplateID string          `json:"template_id"`
	Fields     []PropertyField `json:"fields"`
}

// FileExportInfo specifies export info.
type FileExportInfo struct {
	ExportAs string `json:"export_as"`
}

// FileSymlinkInfo specifies symlink info.
type FileSymlinkInfo struct {
	Target string `json:"target"`
}

const (
	MetadataTypeFile    = "file"
	MetadataTypeFolder  = "folder"
	MetadataTypeDeleted = "deleted"
)

// Metadata for a file or folder.
type Metadata struct {
	Tag                      string           `json:".tag"`
	Name                     string           `json:"name"`
	PathLower                string           `json:"path_lower"`
	PathDisplay              string           `json:"path_display"`
	ClientModified           time.Time        `json:"client_modified"`
	ServerModified           time.Time        `json:"server_modified"`
	Rev                      string           `json:"rev"`
	Size                     uint64           `json:"size"`
	ID                       string           `json:"id"`
	MediaInfo                *MediaInfo       `json:"media_info,omitempty"`
	SymlinkInfo              *FileSymlinkInfo `json:"symlink_info,omitempty"`
	SharingInfo              *FileSharingInfo `json:"sharing_info,omitempty"`
	IsDownloadable           bool             `json:"is_downloadable"`
	ExportInfo               *FileExportInfo  `json:"export_info,omitempty"`
	PropertyGroups           []*PropertyGroup `json:"property_groups,omitempty"`
	HasExplicitSharedMembers bool             `json:"has_explicit_shared_members,omitempty"`
	ContentHash              string           `json:"content_hash,omitempty"`
}

// NewMetadata creates Metadata and set default values.
func NewMetadata() *Metadata {
	return &Metadata{
		IsDownloadable: true,
	}
}

// IsFile returns true if 'm' is file object
func (m *Metadata) IsFile() bool {
	return (strings.ToLower(m.Tag) == MetadataTypeFile)
}

// IsFolder returns true if 'm' is folder object
func (m *Metadata) IsFolder() bool {
	return (strings.ToLower(m.Tag) == MetadataTypeFolder)
}

// IsDeleted returns true if 'm' is deleted object
func (m *Metadata) IsDeleted() bool {
	return (strings.ToLower(m.Tag) == MetadataTypeDeleted)
}

// MetadataV2 metadata for a file, folder or deleted.
type MetadataV2 struct {
	Metadata *Metadata `json:"metadata"`
}

// NewMetadataV2 creates MetadataV2 and set default values.
func NewMetadataV2() *MetadataV2 {
	return &MetadataV2{
		Metadata: NewMetadata(),
	}
}

// TemplateFilterBase specifies template filter base.
type TemplateFilterBase struct {
	FilterSome []string `json:"filter_some"`
}

// NewTemplateFilterBase creates TemplateFilterBase
func NewTemplateFilterBase() *TemplateFilterBase {
	return &TemplateFilterBase{}
}

// GetMetadataInput request input.
type GetMetadataInput struct {
	Path                            string              `json:"path"`
	IncludeMediaInfo                bool                `json:"include_media_info,omitempty"`
	IncludeDeleted                  bool                `json:"include_deleted,omitempty"`
	IncludeHasExplicitSharedMembers bool                `json:"include_has_explicit_shared_members,omitempty"`
	IncludePropertyGroups           *TemplateFilterBase `json:"include_property_groups,omitempty"`
}

// GetMetadataOutput request output.
type GetMetadataOutput struct {
	Metadata
}

// GetMetadata returns the metadata for a file or folder.
func (c *Files) GetMetadata(in *GetMetadataInput) (out *GetMetadataOutput, err error) {
	body, err := c.call("/files/get_metadata", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// CreateFolderInput request input.
type CreateFolderInput struct {
	Path       string `json:"path"`
	AutoRename bool   `json:"autorename,omitempty"`
}

// CreateFolderOutput request output.
type CreateFolderOutput struct {
	MetadataV2
}

// CreateFolder creates a folder.
func (c *Files) CreateFolder(in *CreateFolderInput) (out *CreateFolderOutput, err error) {
	body, err := c.call("/files/create_folder_v2", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// DeleteInput request input.
type DeleteInput struct {
	Path      string `json:"path"`
	ParentRev string `json:"parent_rev,omitempty"`
}

// DeleteOutput request output.
type DeleteOutput struct {
	MetadataV2
}

// Delete a file or folder and its contents.
func (c *Files) Delete(in *DeleteInput) (out *DeleteOutput, err error) {
	body, err := c.call("/files/delete_v2", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// PermanentlyDeleteInput request input.
type PermanentlyDeleteInput struct {
	Path      string `json:"path"`
	ParentRev string `json:"parent_rev,omitempty"`
}

// PermanentlyDelete a file or folder and its contents.
func (c *Files) PermanentlyDelete(in *PermanentlyDeleteInput) (err error) {
	body, err := c.call("/files/permanently_delete", in)
	if err != nil {
		return
	}
	defer body.Close()

	return
}

// CopyInput request input.
type CopyInput struct {
	FromPath               string `json:"from_path"`
	ToPath                 string `json:"to_path"`
	AllowSharedFolder      bool   `json:"allow_shared_folder,omitempty"`
	AutoRename             bool   `json:"autorename,omitempty"`
	AllowOwnershipTransfer bool   `json:"allow_ownership_transfer,omitempty"`
}

// CopyOutput request output.
type CopyOutput struct {
	MetadataV2
}

// Copy a file or folder to a different location.
func (c *Files) Copy(in *CopyInput) (out *CopyOutput, err error) {
	body, err := c.call("/files/copy_v2", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// MoveInput request input.
type MoveInput struct {
	FromPath               string `json:"from_path"`
	ToPath                 string `json:"to_path"`
	AllowSharedFolder      bool   `json:"allow_shared_folder,omitempty"`
	AutoRename             bool   `json:"autorename,omitempty"`
	AllowOwnershipTransfer bool   `json:"allow_ownership_transfer,omitempty"`
}

// MoveOutput request output.
type MoveOutput struct {
	MetadataV2
}

// Move a file or folder to a different location.
func (c *Files) Move(in *MoveInput) (out *MoveOutput, err error) {
	body, err := c.call("/files/move_v2", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// RestoreInput request input.
type RestoreInput struct {
	Path string `json:"path"`
	Rev  string `json:"rev"`
}

// RestoreOutput request output.
type RestoreOutput struct {
	Metadata
}

// Restore a file to a specific revision.
func (c *Files) Restore(in *RestoreInput) (out *RestoreOutput, err error) {
	body, err := c.call("/files/restore", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// ListFolderInput request input.
type ListFolderInput struct {
	Path                            string `json:"path"`
	Recursive                       bool   `json:"recursive"`
	IncludeMediaInfo                bool   `json:"include_media_info"`
	IncludeDeleted                  bool   `json:"include_deleted"`
	IncludeHasExplicitSharedMembers bool   `json:"include_has_explicit_shared_members"`
	IncludeMountedFolders           bool   `json:"include_mounted_folders"`
	IncludeNonDownloadableFiles     bool   `json:"include_non_downloadable_files"`
}

// NewListFolderInput creates new ListFolderInput and set default values
func NewListFolderInput() *ListFolderInput {
	return &ListFolderInput{
		IncludeMountedFolders:       true,
		IncludeNonDownloadableFiles: true,
	}
}

// ListFolderOutput request output.
type ListFolderOutput struct {
	Cursor  string      `json:"cursor"`
	HasMore bool        `json:"has_more"`
	Entries []*Metadata `json:"entries"`
}

// ListFolder returns the metadata for a file or folder.
func (c *Files) ListFolder(in *ListFolderInput) (out *ListFolderOutput, err error) {
	in.Path = normalizePath(in.Path)

	body, err := c.call("/files/list_folder", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// ListFolderContinueInput request input.
type ListFolderContinueInput struct {
	Cursor string `json:"cursor"`
}

// ListFolderContinue pagenates using the cursor from ListFolder.
func (c *Files) ListFolderContinue(in *ListFolderContinueInput) (out *ListFolderOutput, err error) {
	body, err := c.call("/files/list_folder/continue", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// HighlightSpan represents a highlight span.
type HighlightSpan struct {
	HighlightStr  string `json:"highlight_str"`
	IsHighlighted bool   `json:"is_highlighted"`
}

// SearchMatchV2 represents a matched file, folder or deleted.
type SearchMatchV2 struct {
	MetadataV2
	HighlightSpans []*HighlightSpan `json:"highlight_spans"`
}

// FileStatusType file status types.
type FileStatusType string

const (
	FileStatusActive  FileStatusType = "active"
	FileStatusDeleted FileStatusType = "deleted"
)

// FileCategoryType file category types.
type FileCategoryType string

const (
	FileCategoryImage        FileCategoryType = "image"        // jpg, png, gif, and more.
	FileCategoryDocument     FileCategoryType = "document"     // doc, docx, txt, and more.
	FileCategoryPDF          FileCategoryType = "pdf"          // pdf.
	FileCategorySpreadSheet  FileCategoryType = "spreadsheet"  // xlsx, xls, csv, and more.
	FileCategoryPresentation FileCategoryType = "presentation" // ppt, pptx, key, and more.
	FileCategoryAudio        FileCategoryType = "audio"        // mp3, wav, mid, and more.
	FileCategoryVideo        FileCategoryType = "video"        // mov, wmv, mp4, and more.
	FileCategoryFolder       FileCategoryType = "folder"       // dropbox folder.
	FileCategoryPaper        FileCategoryType = "paper"        // dropbox paper doc.
	FileCategoryOthers       FileCategoryType = "others"       // any file not in one of the categories above.
)

// SearchOptions represents search options.
type SearchOptions struct {
	Path           string             `json:"path,omitempty"`
	MaxResults     uint64             `json:"max_results"` // min=1, max=1000
	FileStatus     FileStatusType     `json:"file_status"`
	FilenameOnly   bool               `json:"filename_only,omitempty"`
	FileExtensions []string           `json:"file_extensions,omitempty"`
	FileCategories []FileCategoryType `json:"file_categories,omitempty"`
}

// NewSearchOptions creates new SearchOptions and set default values
func NewSearchOptions() *SearchOptions {
	return &SearchOptions{
		MaxResults: 100,
		FileStatus: FileStatusActive,
	}
}

// SearchInput request input.
type SearchInput struct {
	Query             string         `json:"query"`
	Options           *SearchOptions `json:"options,omitempty"`
	IncludeHighlights bool           `json:"include_highlights,omitempty"`
}

// SearchOutput request output.
type SearchOutput struct {
	Matches []*SearchMatchV2 `json:"matches"`
	HasMore bool             `json:"has_more"`
	Cursor  string           `json:"cursor"`
}

// Search for files and folders.
func (c *Files) Search(in *SearchInput) (out *SearchOutput, err error) {
	if in.Options != nil {
		in.Options.Path = normalizePath(in.Options.Path)
	}

	body, err := c.call("/files/search_v2", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// SearchContinueInput request input.
type SearchContinueInput struct {
	Cursor string `json:"cursor"`
}

// SearchContinue pagenates using the cursor from Search.
func (c *Files) SearchContinue(in *SearchContinueInput) (out *SearchOutput, err error) {
	body, err := c.call("/files/search/continue_v2", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// UploadInput request input.
type UploadInput struct {
	Path           string           `json:"path"`
	Mode           interface{}      `json:"mode,omitempty"`
	AutoRename     bool             `json:"autorename,omitempty"`
	ClientModified string           `json:"client_modified,omitempty"`
	Mute           bool             `json:"mute,omitempty"`
	PropertyGroups []*PropertyGroup `json:"property_groups,omitempty"`
	StrictConflict bool             `json:"strict_conflict,omitempty"`
	Reader         io.Reader        `json:"-"`
}

// NewUploadInput creates UploadInput and set default values.
func NewUploadInput() *UploadInput {
	return &UploadInput{
		Mode: WriteModeAdd,
	}
}

// SetMode sets write mode.
func (i *UploadInput) SetMode(mode WriteMode, rev string) {
	if mode == WriteModeUpdate {
		i.Mode = newWriteModeUpdate(rev)
	} else {
		i.Mode = mode
	}
}

// GetMode gets write mode.
func (i *UploadInput) GetMode() (mode WriteMode, rev string) {
	switch m := i.Mode.(type) {
	case *writeModeUpdate:
		return WriteModeUpdate, m.Rev
	case WriteMode:
		return m, ""
	}

	return "", ""
}

// checkMode checks write mode.
func (i *UploadInput) checkMode() {
	if i.Mode == nil {
		i.SetMode(WriteModeAdd, "")
	}
}

// UploadOutput request output.
type UploadOutput struct {
	Metadata
}

// Upload a file smaller than 150MB.
func (c *Files) Upload(in *UploadInput) (out *UploadOutput, err error) {
	in.checkMode()

	body, _, err := c.download("/files/upload", in, in.Reader)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// DownloadInput request input.
type DownloadInput struct {
	Path string `json:"path"`
}

// DownloadOutput request output.
type DownloadOutput struct {
	Body   io.ReadCloser
	Length int64
}

// Download a file.
func (c *Files) Download(in *DownloadInput) (out *DownloadOutput, err error) {
	body, l, err := c.download("/files/download", in, nil)
	if err != nil {
		return
	}

	out = &DownloadOutput{body, l}
	return
}

// ThumbnailFormat determines the format of the thumbnail.
type ThumbnailFormat string

const (
	// ThumbnailFormatJPEG specifies a JPG thumbnail
	ThumbnailFormatJPEG ThumbnailFormat = "jpeg"
	// ThumbnailFormatPNG specifies a PNG thumbnail
	ThumbnailFormatPNG ThumbnailFormat = "png"
)

// ThumbnailSize determines the size of the thumbnail.
type ThumbnailSize string

const (
	// ThumbnailSizeW32H32 specifies a size of 32 by 32 px
	ThumbnailSizeW32H32 ThumbnailSize = "w32h32"
	// ThumbnailSizeW64H64 specifies a size of 64 by 64 px
	ThumbnailSizeW64H64 ThumbnailSize = "w64h64"
	// ThumbnailSizeW128H128 specifies a size of 128 by 128 px
	ThumbnailSizeW128H128 ThumbnailSize = "w128h128"
	// ThumbnailSizeW256H256 specifies a size of 256 by 256 px
	ThumbnailSizeW256H256 ThumbnailSize = "w256h256"
	// ThumbnailSizeW480H320 specifies a size of 480 by 320 px
	ThumbnailSizeW480H320 ThumbnailSize = "w480h320"
	// ThumbnailSizeW640H480 specifies a size of 640 by 480 px
	ThumbnailSizeW640H480 ThumbnailSize = "w640h480"
	// ThumbnailSizeW960H640 specifies a size of 960 by 640 px
	ThumbnailSizeW960H640 ThumbnailSize = "w960h640"
	// ThumbnailSizeW1024H768 specifies a size of 1024 by 768 px
	ThumbnailSizeW1024H768 ThumbnailSize = "w1024h768"
	// ThumbnailSizeW2048H1536 specifies a size of 2048 by 1536 px
	ThumbnailSizeW2048H1536 ThumbnailSize = "w2048h1536"
)

// ThumbnailMode determines the mode of the thumbnail.
type ThumbnailMode string

const (
	// ThumbnailModeStrict scale down the image to fit within the given size.
	ThumbnailModeStrict ThumbnailMode = "strict"
	// ThumbnailModeBestfit scale down the image to fit within the given size or its transpose.
	ThumbnailModeBestfit ThumbnailMode = "bestfit"
	// ThumbnailModeFitoneBestfit scale down the image to completely cover the given size or its transpose.
	ThumbnailModeFitoneBestfit ThumbnailMode = "fitone_bestfit"
)

// GetThumbnailInput request input.
type GetThumbnailInput struct {
	Path   string          `json:"path"`
	Format ThumbnailFormat `json:"format"`
	Size   ThumbnailSize   `json:"size"`
	Mode   ThumbnailMode   `json:"mode"`
}

// NewGetThumbnailInput creates GetThumbnailInput and set default values.
func NewGetThumbnailInput() *GetThumbnailInput {
	return &GetThumbnailInput{
		Format: ThumbnailFormatJPEG,
		Size:   ThumbnailSizeW64H64,
		Mode:   ThumbnailModeStrict,
	}
}

// GetThumbnailOutput request output.
type GetThumbnailOutput struct {
	Body   io.ReadCloser
	Length int64
}

// GetThumbnail a thumbnail for a file. Currently thumbnails are only generated for the
// files with the following extensions: png, jpeg, png, tiff, tif, gif and bmp.
func (c *Files) GetThumbnail(in *GetThumbnailInput) (out *GetThumbnailOutput, err error) {
	body, l, err := c.download("/files/get_thumbnail", in, nil)
	if err != nil {
		return
	}

	out = &GetThumbnailOutput{body, l}
	return
}

// GetPreviewInput request input.
type GetPreviewInput struct {
	Path string `json:"path"`
}

// GetPreviewOutput request output.
type GetPreviewOutput struct {
	Body   io.ReadCloser
	Length int64
}

// GetPreview a preview for a file. Currently previews are only generated for the
// files with the following extensions: .doc, .docx, .docm, .ppt, .pps, .ppsx,
// .ppsm, .pptx, .pptm, .xls, .xlsx, .xlsm, .rtf
func (c *Files) GetPreview(in *GetPreviewInput) (out *GetPreviewOutput, err error) {
	body, l, err := c.download("/files/get_preview", in, nil)
	if err != nil {
		return
	}

	out = &GetPreviewOutput{body, l}
	return
}

// ListRevisionsMode determines the list revisions mode.
type ListRevisionsMode string

const (
	// ListRevisionsModePath path mode.
	ListRevisionsModePath ListRevisionsMode = "path"
	// ListRevisionsModeID id mode.
	ListRevisionsModeID ListRevisionsMode = "id"
)

// ListRevisionsInput request input.
type ListRevisionsInput struct {
	Path  string            `json:"path"`
	Mode  ListRevisionsMode `json:"mode,omitempty"`
	Limit uint64            `json:"limit,omitempty"`
}

// NewListRevisionsInput creates ListRevisionsInput and set default values.
func NewListRevisionsInput() *ListRevisionsInput {
	return &ListRevisionsInput{
		Mode:  ListRevisionsModePath,
		Limit: 10,
	}
}

// ListRevisionsOutput request output.
type ListRevisionsOutput struct {
	IsDeleted     bool        `json:"is_deleted"`
	Entries       []*Metadata `json:"entries"`
	ServerDeleted *time.Time  `json:"server_deleted"`
}

// ListRevisions gets the revisions of the specified file.
func (c *Files) ListRevisions(in *ListRevisionsInput) (out *ListRevisionsOutput, err error) {
	body, err := c.call("/files/list_revisions", in)
	if err != nil {
		return
	}
	defer body.Close()

	err = json.NewDecoder(body).Decode(&out)
	return
}

// Normalize path so people can use "/" as they expect.
func normalizePath(s string) string {
	if s == "/" {
		return ""
	}
	return s
}

const hashBlockSize = 4 * 1024 * 1024

// ContentHash returns the Dropbox content_hash for a io.Reader.
// See https://www.dropbox.com/developers/reference/content-hash
func ContentHash(r io.Reader) (string, error) {
	buf := make([]byte, hashBlockSize)
	resultHash := sha256.New()
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	if n > 0 {
		bufHash := sha256.Sum256(buf[:n])
		resultHash.Write(bufHash[:])
	}
	for n == hashBlockSize && err == nil {
		n, err = r.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n > 0 {
			bufHash := sha256.Sum256(buf[:n])
			resultHash.Write(bufHash[:])
		}
	}
	return fmt.Sprintf("%x", resultHash.Sum(nil)), nil
}

// FileContentHash returns the Dropbox content_hash for a local file.
// See https://www.dropbox.com/developers/reference/content-hash
func FileContentHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return ContentHash(f)
}
