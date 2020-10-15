package main

import (
	"gitlab.com/CS425_MPs/FileService" // go mod init "gitlab.com/CS425_MPs"
	"golang.org/x/net/context"
	"google.golang.org/grpc" // go get -u google.golang.org/grpc
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"os"
)

var (
	dialOpts = [2]grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}
)

// Server methods

type FileTransferServer struct{}

func (mem *Member) InitializeServer(port string) {
	serverListener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	service.RegisterFileTransferServer(grpcServer, &FileTransferServer{})
	reflection.Register(grpcServer)

	if err2 := grpcServer.Serve(serverListener); err2 != nil {
		panic(err2)
	}
}

func (s *FileTransferServer) Upload(ctx context.Context, uploadReq *service.UploadRequest) (*service.UploadReply, error) {
	file, err := os.Create(uploadReq.SdfsFileName)
	if err != nil {
		return &service.UploadReply{Status: false}, err
	}
	defer file.Close()

	file.Write([]byte(uploadReq.FileContents))

	return &service.UploadReply{Status: true}, nil
}

func (s *FileTransferServer) Download(ctx context.Context, downloadReq *service.DownloadRequest) (*service.DownloadReply, error) {
	file, err := os.Open(downloadReq.GetSdfsFileName())
	defer file.Close()
	if err != nil {
		return &service.DownloadReply{DoesFileExist: false, FileContents: err.Error()}, nil
	}

	buf := make([]byte, 1024)
	size, err := file.Read(buf)
	if err != nil {
		return &service.DownloadReply{DoesFileExist: false, FileContents: err.Error()}, nil
	}

	fileStr := string(buf[:size])
	return &service.DownloadReply{DoesFileExist: true, FileContents: fileStr}, nil
}

// Client Methods

func GetFileContents(localFileName string) string {
	content, err := ioutil.ReadFile(localFileName)
	if err != nil {
		Warn.Println("Unable to read file.")
		return ""
	}

	// Convert []byte to string
	return string(content)
}

func (mem *Member) Upload(ipAddr string, port string, localFileName string, sdfsFileName string) {
	dest := ipAddr + ":" + port
	conn, err := grpc.Dial(dest, dialOpts[0:2]...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := service.NewFileTransferClient(conn)

	fileContents := GetFileContents(localFileName)
	uploadReply, err2 := client.Upload(context.Background(), &service.UploadRequest{
		FileContents: fileContents,
		SdfsFileName: sdfsFileName})
	if err2 != nil {
		Warn.Println("fail to upload: ", err2)
		return
	}

	if uploadReply.GetStatus() == true {
		Info.Println("Successfully uploaded file: [", localFileName, "] as [", sdfsFileName, "] at addr ", dest)
	}
}

func (mem *Member) Download(ipAddr string, port string, sdfsFileName string, localFileName string) {
	dest := ipAddr + ":" + port
	conn, err := grpc.Dial(dest, dialOpts[0:2]...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := service.NewFileTransferClient(conn)

	downloadReply, err2 := client.Download(context.Background(),
		&service.DownloadRequest{SdfsFileName: sdfsFileName})

	if !downloadReply.DoesFileExist {
		Warn.Println("Failed to download ", sdfsFileName, ". File does not exist.")
		return
	}

	file, err2 := os.Create(localFileName)
	if err2 != nil {
		Warn.Println("Failed to create file: ", err2)
		return
	}
	defer file.Close()

	file.Write([]byte(downloadReply.FileContents))
	Info.Println("Successfully downloaded file: [", sdfsFileName, "]")
}
