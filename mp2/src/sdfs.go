package main

import (
	"errors"
	"gitlab.com/CS425_MPs/FileService" // go mod init "gitlab.com/CS425_MPs"
	"golang.org/x/net/context"
	"google.golang.org/grpc" // go get -u google.golang.org/grpc
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"
)

var (
	// 1346378950 is the size of wiki corpus + some more for fun lol
	dialSize          = 1346378950 + 2048
	uploadChunkSize   = 350000
	downloadChunkSize = 10000000
	clientDialOpts    = [4]grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(dialSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(dialSize)),
		grpc.WithReturnConnectionError()}
	serverDialOpts        = [2]grpc.ServerOption{grpc.MaxRecvMsgSize(dialSize), grpc.MaxSendMsgSize(dialSize)}
	dirName        string = "SDFS"
)

// Init
func InitSdfsDirectory() {
	_, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		os.MkdirAll(dirName, 0755)
	} else {
		// clear contents when starting up
		dir, _ := ioutil.ReadDir(dirName)
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{dirName, d.Name()}...))
		}
	}
}

// Server methods

type FileTransferServer struct{}

func InitializeServer(port string) {
	serverListener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(serverDialOpts[0:2]...)
	service.RegisterFileTransferServer(grpcServer, &FileTransferServer{})
	reflection.Register(grpcServer)

	if err2 := grpcServer.Serve(serverListener); err2 != nil {
		panic(err2)
	}
}

func (s *FileTransferServer) Upload(ctx context.Context, uploadReq *service.UploadRequest) (*service.UploadReply, error) {
	filePath := filepath.Join(dirName, filepath.Base(uploadReq.SdfsFileName))
	fileFlags := os.O_CREATE | os.O_WRONLY
	if uploadReq.IsMultipleChunks && !uploadReq.IsFirstChunk {
		fileFlags = fileFlags | os.O_APPEND
	}

	file, err := os.OpenFile(filePath, fileFlags, 0777)
	if err != nil {
		return &service.UploadReply{Status: false}, err
	}
	defer file.Close()

	file.Write(uploadReq.FileContents)

	return &service.UploadReply{Status: true}, nil
}

func (s *FileTransferServer) Download(ctx context.Context, downloadReq *service.DownloadRequest) (*service.DownloadReply, error) {
	file, err := os.Open(filepath.Join(dirName, filepath.Base(downloadReq.GetSdfsFileName())))
	defer file.Close()
	if err != nil {
		return &service.DownloadReply{
			DoesFileExist:    false,
			FileContents:     []byte(err.Error()),
			IsMultipleChunks: false,
			IsLastChunk:      true}, nil
	}

	fileStat, err2 := file.Stat()
	if err2 != nil {
		return &service.DownloadReply{
			DoesFileExist:    false,
			FileContents:     []byte(err.Error()),
			IsMultipleChunks: false,
			IsLastChunk:      true}, nil
	}

	fileSize := fileStat.Size()
	isMultChunks := false
	if fileSize >= int64(downloadChunkSize) {
		isMultChunks = true
	}

	isLastChunk := true
	// + 1 to know how many are about to be sent
	if fileSize > (int64(downloadReq.ChunkNum+1) * int64(downloadChunkSize)) {
		isLastChunk = false
	}

	// move to position you want to read from
	startIdx := int64(downloadReq.ChunkNum) * int64(downloadChunkSize)
	readSize := int64(downloadChunkSize)
	if isLastChunk {
		readSize = fileSize - startIdx
	}

	buf := make([]byte, readSize)

	var whence int = 0
	_, seekErr := file.Seek(startIdx, whence)
	if seekErr != nil {
		return &service.DownloadReply{
			DoesFileExist:    false,
			FileContents:     []byte(err.Error()),
			IsMultipleChunks: false,
			IsLastChunk:      true}, nil
	}

	file.Read(buf)

	return &service.DownloadReply{
		DoesFileExist:    true,
		FileContents:     buf,
		IsMultipleChunks: isMultChunks,
		IsLastChunk:      isLastChunk}, nil
}

// Client Methods

func DialServer(dest string) (*grpc.ClientConn, error) {
	connectChan := make(chan bool, 1)
	var conn *grpc.ClientConn
	var connErr error
	go func() {
		conn, connErr = grpc.Dial(dest, clientDialOpts[0:4]...)
		connectChan <- true
	}()

	select {
	case <-connectChan:
		Info.Println("Connected to ", dest, " to upload.")
	case <-time.After(time.Duration(Configuration.Settings.failTimeout) * time.Second):
		errorMsg := "Time to connect has surpassed deadline."
		Warn.Println(errorMsg)
		return nil, errors.New(errorMsg)
	}

	if connErr != nil {
		panic(connErr)
	}

	return conn, connErr
}

func GetFileContents(localFileName string) []byte {
	content, err := ioutil.ReadFile(localFileName)
	if err != nil {
		Warn.Println("Unable to read file.")
		return []byte{}
	}

	// Convert []byte to string
	return content
}

func UploadFile(conn *grpc.ClientConn, dest string, fileChunk []byte,
	sdfsFileName string, isMultChunks bool, isFirstChunk bool) error {

	client := service.NewFileTransferClient(conn)
	uploadReply, err2 := client.Upload(context.Background(), &service.UploadRequest{
		FileContents:     fileChunk,
		SdfsFileName:     sdfsFileName,
		IsMultipleChunks: isMultChunks,
		IsFirstChunk:     isFirstChunk})
	if err2 != nil {
		Warn.Println(err2)
		return err2
	}

	if uploadReply.GetStatus() == true {
		Info.Println("Successfully uploaded chunk: [", sdfsFileName, "] at addr ", dest)
		return nil
	}

	errorMsg := "Error: Bad reply status."
	Warn.Println(errorMsg)
	return errors.New(errorMsg)
}

func Upload(ipAddr string, port string, localFileName string, sdfsFileName string) error {
	dest := ipAddr + ":" + port
	conn, connErr := DialServer(dest)
	if connErr != nil {
		return connErr
	}
	defer conn.Close()

	fileContents := GetFileContents(localFileName)
	fileSize := len(fileContents)
	isMultChunks := false
	isFirstChunk := true
	if fileSize >= uploadChunkSize {
		isMultChunks = true
	}

	for i := 0; i < fileSize; i += uploadChunkSize {
		lastIdx := i + uploadChunkSize
		if lastIdx > fileSize {
			lastIdx = fileSize
		}

		err := UploadFile(conn, dest, fileContents[i:lastIdx], sdfsFileName, isMultChunks, isFirstChunk)
		if err != nil {
			return err
		}

		if isFirstChunk {
			isFirstChunk = false
		}

		// sleep so that other threads can wake up
                /*
		if isMultChunks {
			time.Sleep(4 * time.Millisecond)
		}*/
	}

	return nil
}

func DownloadFile(filePath string, fileChunk []byte, fileFlags int) error {
	file, err := os.OpenFile(filePath, fileFlags, 0777)
	if err != nil {
		errorMsg := "Failed to create file."
		Warn.Println(errorMsg, err)
		return errors.New(errorMsg)
	}
	defer file.Close()

	file.Write(fileChunk)

	return nil
}

func Download(ipAddr string, port string, sdfsFileName string, localFileName string) error {
	// establish connection with server
	dest := ipAddr + ":" + port
	conn, connErr := DialServer(dest)
	if connErr != nil {
		return connErr
	}
	defer conn.Close()

	client := service.NewFileTransferClient(conn)

	// get first chunk
	chunkNum := 0
	downloadReply, err2 := client.Download(context.Background(),
		&service.DownloadRequest{SdfsFileName: sdfsFileName, ChunkNum: int32(chunkNum)})

	if err2 != nil || !downloadReply.DoesFileExist {
		errorMsg := "Error: Unable to download file " + sdfsFileName + ". File does not exist."
		Warn.Println(errorMsg, err2)
		return errors.New(errorMsg)
	}

	fileFlags := os.O_CREATE | os.O_WRONLY

	for {
		// save reply contents to file path
		dlErr := DownloadFile(localFileName, downloadReply.FileContents, fileFlags)
		if dlErr != nil {
			return dlErr
		}

		if downloadReply.IsLastChunk {
			break
		}

		// if here, we're gonna append to the file
		fileFlags = fileFlags | os.O_APPEND

		// sleep before requesting next chunk so that other threads can run lol
		time.Sleep(4 * time.Millisecond)

		// get next chunk
		chunkNum += 1
		downloadReply, err2 = client.Download(context.Background(),
			&service.DownloadRequest{SdfsFileName: sdfsFileName, ChunkNum: int32(chunkNum)})

		if err2 != nil {
			Warn.Println("Error in download process.")
			return err2
		}
	}

	Info.Println("Successfully downloaded file: [", sdfsFileName, "]")
	return nil
}
