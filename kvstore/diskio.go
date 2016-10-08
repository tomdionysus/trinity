package kvstore

import (
	"syscall"
	// "errors"
)

type DiskIO struct {
	DeviceName string
	BlockSize  uint

	fileDescriptor int
}

func NewDiskIO(deviceName string, blocksize uint) *DiskIO {
	inst := &DiskIO{
		DeviceName: deviceName,
		BlockSize:  blocksize,
	}
	return inst
}

func (dio *DiskIO) Open() error {
	fd, err := syscall.Open(dio.DeviceName, syscall.O_RDWR, 0777)
	if err != nil {
		return err
	}
	dio.fileDescriptor = fd
	return nil
}

func (dio *DiskIO) Close() error {
	return syscall.Close(dio.fileDescriptor)
}

func (dio *DiskIO) ReadBlock(blockaddr uint64, buffer []byte) (uint, error) {
	var noff int64 = int64(blockaddr) * int64(dio.BlockSize)
	_, err := syscall.Seek(dio.fileDescriptor, noff, 0)
	if err != nil {
		return 0, err
	}
	num, err := syscall.Read(dio.fileDescriptor, buffer)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}

func (dio *DiskIO) WriteBlock(blockaddr uint64, buffer []byte) (uint, error) {
	var noff int64 = int64(blockaddr) * int64(dio.BlockSize)
	_, err := syscall.Seek(dio.fileDescriptor, noff, 0)
	if err != nil {
		return 0, err
	}
	num, err := syscall.Write(dio.fileDescriptor, buffer)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}
