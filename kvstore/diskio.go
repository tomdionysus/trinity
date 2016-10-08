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

func (me *DiskIO) Open() error {
	fd, err := syscall.Open(me.DeviceName, syscall.O_RDWR, 0777)
	if err != nil {
		return err
	}
	me.fileDescriptor = fd
	return nil
}

func (me *DiskIO) Close() error {
	return syscall.Close(me.fileDescriptor)
}

func (me *DiskIO) ReadBlock(blockaddr uint64, buffer []byte) (uint, error) {
	var noff int64 = int64(blockaddr) * int64(me.BlockSize)
	_, err := syscall.Seek(me.fileDescriptor, noff, 0)
	if err != nil {
		return 0, err
	}
	num, err := syscall.Read(me.fileDescriptor, buffer)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}

func (me *DiskIO) WriteBlock(blockaddr uint64, buffer []byte) (uint, error) {
	var noff int64 = int64(blockaddr) * int64(me.BlockSize)
	_, err := syscall.Seek(me.fileDescriptor, noff, 0)
	if err != nil {
		return 0, err
	}
	num, err := syscall.Write(me.fileDescriptor, buffer)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}
