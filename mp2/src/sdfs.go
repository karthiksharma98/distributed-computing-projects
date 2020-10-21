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
	"path/filepath"
	"time"
)

var (
	// 1346378950 is the size of wiki corpus + some more for fun lol
	dialSize       = 1346378950 + 2048
	clientDialOpts = [4]grpc.DialOption{grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(dialSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(dialSize))}
	serverDialOpts        = [2]grpc.ServerOption{grpc.MaxRecvMsgSize(dialSize), grpc.MaxSendMsgSize(dialSize)}
	dirName        string = "SDFS"
)

// Init
func InitSdfsDirectory() {
	_, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		os.MkdirAll(dirName, 0755)
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
		return &service.DownloadReply{DoesFileExist: false, FileContents: []byte(err.Error())}, nil
	}

	fileStat, err2 := file.Stat()
	if err2 != nil {
		return &service.DownloadReply{DoesFileExist: false, FileContents: []byte(err2.Error())}, nil
	}

	buf := make([]byte, fileStat.Size())
	file.Read(buf)
	return &service.DownloadReply{DoesFileExist: true, FileContents: buf}, nil
}

// Client Methods

func GetFileContents(localFileName string) []byte {
	content, err := ioutil.ReadFile(localFileName)
	if err != nil {
		Warn.Println("Unable to read file.")
		return []byte{}
	}

	// Convert []byte to string
	return content
}

func DialAndSend(conn *grpc.ClientConn, dest string, fileChunk []byte,
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
	conn, err := grpc.Dial(dest, clientDialOpts[0:4]...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fileContents := GetFileContents(localFileName)
	fileSize := len(fileContents)
	chunkSize := 350000
	isMultChunks := false
	isFirstChunk := true
	if fileSize >= chunkSize {
		isMultChunks = true
	}

	for i := 0; i < fileSize; i += chunkSize {
		lastIdx := i + chunkSize
		if lastIdx > fileSize {
			lastIdx = fileSize
		}

		fileChunk := fileContents[i:lastIdx]
		err := DialAndSend(conn, dest, fileChunk, sdfsFileName, isMultChunks, isFirstChunk)
		if err != nil {
			return err
		}

		if isFirstChunk {
			isFirstChunk = false
		}

		// sleep so that other threads can wake up
		if isMultChunks {
			time.Sleep(4 * time.Millisecond)
		}
	}

	return nil
}

func Download(ipAddr string, port string, sdfsFileName string, localFileName string) error {
	dest := ipAddr + ":" + port
	conn, err := grpc.Dial(dest, clientDialOpts[0:4]...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := service.NewFileTransferClient(conn)

	downloadReply, err2 := client.Download(context.Background(),
		&service.DownloadRequest{SdfsFileName: sdfsFileName})

	if !downloadReply.DoesFileExist {
		errorMsg := "Error: Unable to download file " + sdfsFileName + ". File does not exist."
		Warn.Println(errorMsg, err2)
		return errors.New(errorMsg)
	}

	file, err2 := os.Create(localFileName)
	if err2 != nil {
		errorMsg := "Failed to create file."
		Warn.Println(errorMsg, err2)
		return errors.New(errorMsg)
	}
	defer file.Close()

	file.Write(downloadReply.FileContents)
	Info.Println("Successfully downloaded file: [", sdfsFileName, "]")
	return nil
}
