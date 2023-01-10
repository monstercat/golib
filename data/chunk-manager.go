package data

import (
	"io"
	"math"
	"strconv"
	"sync"
	"time"
)

// ChunkManager manages uploading chunks by providing a goroutine which performs the chunking. It contains a service
// that should be run on a goroutine.
//
// Usage:
//
//	// Initialize
//	manager := &ChunkManager{...}
//	go manager.Run()
//	defer manager.Shutdown()
//
//	// Send a first chunk
//	notifier := manager.PutWithStatus("some filepath", bytes.NewReader(b))
//	for n := range notifier {
//		handleNotifications(n)
//	}
//
//	// To send a second chunk
//	notifier := manager.ResumePutWithStatus("some filepath", bytes.NewReader(b))
//	for n := range notifier {
//		handleNotifications(n)
//	}
//
// The notifier is used to provide updates on the progress of the upload. The notifier returned from StartUpload is the
// same as the one from ResumeUpload. Ensure to close the notifier once done.
//
// The notifier will send a notification of type UploadStatus which contains a code, and either a message or an error.
// In the case that Code is "Error", it will populate the [Error] field. Otherwise, it will populate the `Message field.
//
// When complete, the `UploadStatus` will return a message saying `Upload Complete`.
//
// Implementation Notes:
// =====================
// The ChunkManager only manages the *reception* of data for a ChunkUpload. The actual logic for sending data in chunks
// is handled by the ChunkUpload. Consequently, the ChunkManager only registers the beginning and ending of a chunk as
// well as free data when an incomplete chunk is too old.
type ChunkManager struct {
	// ChunkSizeLimit defines the maximum size a chunk can be.
	ChunkSizeLimit int

	// IncompleteUploadExpiry defines the expiry duration of a chunk.
	IncompleteUploadExpiry time.Duration

	// Creates FileServices that is required to run the ChunkManager. Each file service should handle one upload.
	// For example, if an upload ID is required to sync between chunks, the implementation of ChunkFileService which
	// is returned can store the required upload ID.
	FS ChunkFileServiceFactory

	// Mutex for thread safety.
	lock sync.RWMutex

	// List of 'incomplete' uploads.
	incomplete map[string]*ChunkUpload

	// Shutdown signal
	shutdown chan bool

	// Channel for sending uploads to the chunk manager.
	uploads chan *ChunkUpload
}

func (m *ChunkManager) Init() {
	if m.shutdown == nil {
		m.shutdown = make(chan bool)
	}
	if m.incomplete == nil {
		m.incomplete = make(map[string]*ChunkUpload)
	}
	if m.uploads == nil {
		// We are giving a length of 20 to uploads to delay deadlocking.
		// However, users *must* take note to run "Run" to refresh the upload queue.
		m.uploads = make(chan *ChunkUpload, 20)
	}
}

// Shutdown sends a signal to exit the goroutine.
func (m *ChunkManager) Shutdown() {
	// Recover from the panic in case shutdown is nil .
	defer func() {
		_ = recover()
	}()

	// Only the main process needs to be shutdown. The other goroutines should
	// shut down on their own as they are fitted with exit on timeout.
	close(m.shutdown)
}

// UploadExists returns true if the upload exists.
func (m *ChunkManager) UploadExists(filepath string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	_, ok := m.incomplete[filepath]
	return ok
}

// PutWithStatus tells the ChunkManager to start the upload. It adds the upload to the upload queue and proceeds to
// send data from the provided io.Reader.
func (m *ChunkManager) PutWithStatus(filepath string, filesize int, r io.Reader) chan UploadStatus {
	m.Init()

	// Calculate the number of chunks
	chunks := int(math.Ceil(float64(filesize) / float64(m.ChunkSizeLimit)))

	upload := &ChunkUpload{
		Filepath: filepath,
		FS:       m.FS(),
		Expiry:   time.Now().Add(m.IncompleteUploadExpiry),
		notifier: make(chan UploadStatus, 2),
		parts:    make(chan []byte, chunks),
		Chunks:   chunks,
	}
	m.uploads <- upload
	go upload.Send(r, m.ChunkSizeLimit)
	return upload.notifier
}

// ResumePutWithStatus tells the ChunkManager to resume the upload by sending more data. It assumes the data in StartUpload
// was not enough to complete the upload.
func (m *ChunkManager) ResumePutWithStatus(filepath string, r io.Reader) (chan UploadStatus, error) {
	m.lock.RLock()
	upload := m.incomplete[filepath]
	m.lock.RUnlock()

	// Check if the upload is still processing (it should be). If not, we should not be trying to resume it.
	if !upload.IsProcessing() {
		return nil, ErrUploadNotFound
	}

	// Update the expiry date as we have new data.
	upload.UpdateExpiry(time.Now().Add(m.IncompleteUploadExpiry))
	go upload.Send(r, m.ChunkSizeLimit)
	return upload.notifier, nil
}

// RunUploader starts the goroutine that actually performs the upload for a chunk. Note that this is an infinite loop
// and so should be run in a goroutine.
//
// e.g.,
//
//	go manager.RunUploader()
func (m *ChunkManager) RunUploader() {
	m.Init()
	for {
		select {
		// Exit when we shut down.
		case <-m.shutdown:
			return

		// Every hour, we want to check the expiry of uploads. If the uploads have expired,
		// we want to remove them.
		case <-time.After(time.Hour):
			m.lock.Lock()
			for id, u := range m.incomplete {
				if u.Expiry.After(time.Now()) {
					// Cleanup whatever is required
					u.FS.Cleanup()

					// Delete the data.
					delete(m.incomplete, id)
				}
			}
			m.lock.Unlock()

		// We need to tell the upload to run.
		case upload := <-m.uploads:
			go func() {
				// Register the incomplete upload
				m.lock.Lock()
				m.incomplete[upload.Filepath] = upload
				m.lock.Unlock()

				// Run the upload. Note that, upon error, we *aren't* clearing the `incomplete` cache. This ie because
				// we want to be able to retrieve the upload for resumption.
				if err := upload.Run(); err != nil {
					upload.notifier <- ErrUploadStatus(err)
					return
				}

				// Remove from incomplete.
				m.lock.Lock()
				delete(m.incomplete, upload.Filepath)
				m.lock.Unlock()
			}()
		}
	}
}

// GetIncompleteUpload returns the incomplete upload. Required for ChunkService.
func (m *ChunkManager) GetIncompleteUpload(filepath string) Upload {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.incomplete[filepath]
}

// GetChunkSize returns the chunk size. Required for ChunkService.
func (m *ChunkManager) GetChunkSize() int {
	return m.ChunkSizeLimit
}

// ChunkFileServiceFactory creates a new ChunkFileService for each upload.
type ChunkFileServiceFactory func() ChunkFileService

// ChunkFileService is the service that actually performs the upload operations to a third party (e.g., local file
// system, S3, GCS).
type ChunkFileService interface {
	// IsInitialized should return true if the ChunkFileService has already been initialized.
	IsInitialized() bool

	// InitializeChunk initializes the chunk upload operation.
	InitializeChunk(u *ChunkUpload) error

	// UploadPart should upload the provided bytes to the cloud. Uploaded parts should be completely sequential.
	UploadPart([]byte) error

	// CompleteUpload should perform any necessary steps to finalize the multipart upload.
	CompleteUpload() error

	// Cleanup should delete the ChunkFileService and do any other cleanup necessary. It is only called when an
	// incomplete upload has timed out.
	Cleanup()
}

type InitError struct {
	error error
}

func (e InitError) Error() string {
	return "upload initialization error: " + e.error.Error()
}

type UploadError struct {
	error error
}

func (e UploadError) Error() string {
	return "upload part error: " + e.error.Error()
}

// ChunkUpload is a specific upload related to the chunk.
type ChunkUpload struct {
	// Filepath for the upload.
	Filepath string

	// FileService for running chunks.
	FS ChunkFileService

	// The expiry time for the upload. If the upload has expired, then we should not perform any further processing
	// on it.
	Expiry time.Time

	// Number of chunks for this upload.
	Chunks int

	// Number of received parts.
	numReceivedParts int

	// Number of total upload bytes. This is to provide a progress message giving them number of uploaded bytes.
	uploadedBytes int

	// IsProcessing is a flag that says that the `Run` function is still processing the upload (e.g., the upload
	// is not complete yet). This allows further calls to *SendUpload* to resume an upload.
	isProcessing bool

	// Lock for thread safety
	lock sync.RWMutex

	// We use this to send upload status data back to the user.
	notifier chan UploadStatus

	// A channel to send the data from the reader into the goroutine which actually performs the upload using the
	// provided FS
	parts chan []byte

	// A channel to stop the loading of parts if the incomplete upload has expired.
	shutdown chan bool
}

// GetFilepath returns the filepath. Required to implement Upload interface.
func (u *ChunkUpload) GetFilepath() string {
	return u.Filepath
}

// UpdateExpiry allows updating of the expiry date, for example, when the upload is continued.
func (u *ChunkUpload) UpdateExpiry(tm time.Time) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.Expiry = tm
}

// NumReceivedParts returns the number of parts already received for the chunk upload.
func (u *ChunkUpload) NumReceivedParts() int {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.numReceivedParts
}

// IsComplete returns true when the # of parts received is the same as the expected number of chunks.
func (u *ChunkUpload) IsComplete() bool {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.numReceivedParts >= u.Chunks
}

// incrementNumReceivedParts increments the number of parts
func (u *ChunkUpload) incrementNumReceivedParts() {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.numReceivedParts++
}

// IsProcessing returns true if the required number of chunks hasn't been reached.
func (u *ChunkUpload) IsProcessing() bool {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.isProcessing
}

func (u *ChunkUpload) SetIsProcessing(b bool) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.isProcessing = b
}

// GetUploaded returns the number of bytes already uplaoded. Required to implement Upload interface.
func (u *ChunkUpload) GetUploaded() int {
	u.lock.RLock()
	defer u.lock.RUnlock()

	return u.uploadedBytes
}

func (u *ChunkUpload) UpdateUploadedBytes(newPartBytes int) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.uploadedBytes += newPartBytes
}

// Send sends the data to the goroutine that performs the actual upload in chunks conforming to the chunkSizeLimit.
func (u *ChunkUpload) Send(r io.Reader, chunkSizeLimit int) {
	for {
		b := make([]byte, chunkSizeLimit)
		n, err := r.Read(b)
		if err == io.EOF {
			// do nothing!
			return
		}
		if err != nil {
			u.notifier <- ErrUploadStatus(err)
			return
		}
		if n == 0 {
			continue
		}
		// The byte array *always* has size chunkSizeLimit. However, some parts may have a smaller actual size, so
		// we only want to send the bytes that are read (until n).
		u.parts <- b[:n]
	}
}

// Run performs the logic for this specific upload, ensuring that all upload items have been provided.
func (u *ChunkUpload) Run() error {
	u.SetIsProcessing(true)
	defer func() {
		u.SetIsProcessing(false)
		close(u.parts)
	}()

	if !u.FS.IsInitialized() {
		if err := u.FS.InitializeChunk(u); err != nil {
			// Errors are returned to the calling functions through the notifier channel.
			// returning error here allows for the upload to be sent back to incomplete status.
			//
			// In this case, it shouldn't be sent back to incomplete, as the upload never even started.
			u.notifier <- ErrUploadStatus(InitError{err})
			return nil
		}
	}

	// For each received part, we want to load the part
	u.shutdown = make(chan bool)
	for {
		select {
		case <-u.shutdown:
			return nil
		case part := <-u.parts:
			// Increment the number of received parts.
			u.incrementNumReceivedParts()
			if err := u.FS.UploadPart(part); err != nil {
				u.notifier <- UploadStatus{
					Code:  UploadStatusCodeError,
					Error: err,
				}
				return UploadError{err}
			}

			// Send a status message indicating that
			u.UpdateUploadedBytes(len(part))
			u.notifier <- UploadStatus{
				Code:    UploadStatusCodeProgress,
				Message: strconv.Itoa(u.GetUploaded()),
			}

			// If not complete, wait for the next part.
			if !u.IsComplete() {
				continue
			}

			// When complete, the file service might want to do something.
			if err := u.FS.CompleteUpload(); err != nil {
				u.notifier <- UploadStatus{
					Code:  UploadStatusCodeError,
					Error: err,
				}
				return UploadError{err}
			}

			// Send a message that the upload is complete.
			u.notifier <- UploadStatus{
				Code:    UploadStatusCodeOk,
				Message: "Upload completed",
			}
			return nil
		}
	}
}

func (u *ChunkUpload) Shutdown() {
	// Recover from the panic in case shutdown is nil .
	defer func() {
		_ = recover()
	}()
	close(u.shutdown)
}
