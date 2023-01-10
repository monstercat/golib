package s3util

import (
	"io/ioutil"
	"math"
	"os"
	"testing"
	"time"

	"github.com/monstercat/golib/data"
)

var (
	AwsId  = os.Getenv("AWS_ACCESS_KEY_ID")
	AwsKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	Bucket = os.Getenv("S3_BUCKET")
	Region = os.Getenv("AWS_REGION")

	service = makeService()
)

func makeService() *Service {
	service := &Service{
		Bucket:      Bucket,
		Region:      Region,
		Timeout:     50 * time.Second,
		Concurrency: 1,
		ChunkManager: data.ChunkManager{
			ChunkSizeLimit: DefaultChunkSizeLimit,
			//MinUploadSizeChunked:   DefaultChunkSizeLimit,
			IncompleteUploadExpiry: time.Hour * 24,
		},
	}
	service.ChunkManager.FS = func() data.ChunkFileService {
		return &ChunkFileService{
			Bucket: Bucket,
			Client: service.Client,
		}
	}
	return service
}

const (
	TestFilename1 = "test-file-1.txt"
	TestFilesize1 = 1024 * 1024
	TestFilename2 = "test-file-2.txt"
	TestFilesize2 = 1024 * 1024 * 160
)

func makeFakeFile(filename string, numBytes int) {
	f, err := os.Create("./" + filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	bytes := make([]byte, 0, numBytes)
	for i := 0; i < numBytes; i++ {
		bytes = append(bytes, byte(i))
	}
	if _, err := f.Write(bytes); err != nil {
		panic(err)
	}
}

func cleanupFile(file string) {
	os.Remove("./" + file)
}

func initService() {
	if err := service.Init(AwsId, AwsKey); err != nil {
		panic(err)
	}
	go service.RunUploader()
}

func TestStandardS3Services(t *testing.T) {
	initService()
	defer service.Shutdown()

	makeFakeFile(TestFilename1, TestFilesize1)
	defer cleanupFile(TestFilename1)

	// Upload the small file to FS
	f, err := os.Open("./" + TestFilename1)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := service.Upload(TestFilename1, TestFilesize1, f); err != nil {
		t.Fatal(err)
	}
	exists, err := service.Exists(TestFilename1)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("File should be put, therefore, should exist")
	}

	// Download the file from FS
	rf, err := service.Get(TestFilename1)
	if err != nil {
		t.Fatal(err)
	}

	fileBytes, err := ioutil.ReadAll(rf)
	if err != nil {
		t.Fatal(err)
	}
	if len(fileBytes) != TestFilesize1 {
		t.Fatal("File sizes not the same")
	}

	// Delete the file from FS
	if err := service.Delete(TestFilename1); err != nil {
		t.Fatal(err)
	}
	exists, err = service.Exists(TestFilename1)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("File should be deleted, therefore, should not exist")
	}

	for i, b := range fileBytes {
		if b != byte(i) {
			t.Fatalf("Character is not the same. %d on AWS vs %d on filesystem", b, byte(i))
		}
	}
}

func TestChunkedUpload(t *testing.T) {
	initService()
	defer service.Shutdown()

	// File size is large! This is because we need to test the chunked file uploader.
	makeFakeFile(TestFilename2, TestFilesize2)
	defer cleanupFile(TestFilename2)

	f, err := os.Open("./" + TestFilename2)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	notifier := service.PutWithStatus(TestFilename2, TestFilesize2, f)

	numProgress := 0
L:
	for {
		select {
		case <-time.After(time.Minute * 5):
			t.Fatal("timeout")
		case status := <-notifier:
			switch status.Code {
			case data.UploadStatusCodeOk:
				break L
			case data.UploadStatusCodeError:
				t.Fatal(status.Error)
				return
			case data.UploadStatusCodeProgress:
				numProgress++
			}
		}
	}

	if numProgress != int(math.Ceil(float64(TestFilesize2)/float64(DefaultChunkSizeLimit))) {
		t.Fatal("Expected number of chunks not equal")
	}

	// Retrieve the file and check it!
	rf, err := service.Get(TestFilename2)
	if err != nil {
		t.Fatal(err)
	}
	fileBytes, err := ioutil.ReadAll(rf)
	if err != nil {
		t.Fatal(err)
	}
	if len(fileBytes) != TestFilesize2 {
		t.Error("Files are not the same ")
	}

	for i, b := range fileBytes {
		if b != byte(i) {
			t.Fatalf("Character is not the same. %d on AWS vs %d on filesystem", b, byte(i))
		}
	}

}

func TestResumeChunkUpload(t *testing.T) {
	initService()
	defer service.Shutdown()

	// File size is large! This is because we need to test the chunked file uploader.
	makeFakeFile(TestFilename2, DefaultChunkSizeLimit)
	defer cleanupFile(TestFilename2)

	f, err := os.Open("./" + TestFilename2)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	notifier := service.PutWithStatus(TestFilename2, TestFilesize2, f)

L:
	for {
		select {
		case <-time.After(time.Minute * 5):
			t.Fatal("timeout")
		case status := <-notifier:
			switch status.Code {
			case data.UploadStatusCodeOk:
				t.Fatal("Expecting a timeout error!")
			case data.UploadStatusCodeError:
				break L
			}
		}
	}

	makeFakeFile(TestFilename2+"-part2", TestFilesize2-DefaultChunkSizeLimit)
	defer cleanupFile(TestFilename2 + "-part2")

	f, err = os.Open("./" + TestFilename2 + "-part2")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// We need to now resume the upload!
	f.Seek(0, 0)
	notifier, err = service.ResumePutWithStatus(TestFilename2, f)
	if err != nil {
		t.Fatal(err)
	}

L2:
	for {
		select {
		case <-time.After(time.Minute * 5):
			t.Fatal("timeout")
		case status := <-notifier:
			switch status.Code {
			case data.UploadStatusCodeOk:
				// We should be good!
				break L2
			case data.UploadStatusCodeError:
				t.Fatal(status.Error)
				return
			}
		}
	}

	// Retrieve the file and check it!
	rf, err := service.Get(TestFilename2)
	if err != nil {
		t.Fatal(err)
	}
	fileBytes, err := ioutil.ReadAll(rf)
	if err != nil {
		t.Fatal(err)
	}
	if len(fileBytes) != TestFilesize2 {
		t.Error("Files are not the same ")
	}
	for i, b := range fileBytes {
		if b != byte(i) {
			t.Fatalf("Character is not the same. %d on AWS vs %d on filesystem", b, byte(i))
		}
	}
}
