package fatfs

import (
	"unsafe"

	"modernc.org/libc"
)

func Mount(tls *libc.TLS, fs *FATFS, path string, opt byte) FRESULT {
	_fs := (uintptr)(unsafe.Pointer(fs))
	_path, _ := libc.CString(path)
	return f_mount(tls, _fs, _path, opt)
}

func Open(tls *libc.TLS, fp *FIL, path string, mode uint8) FRESULT {
	_fp := (uintptr)(unsafe.Pointer(fp))
	_path, _ := libc.CString(path)
	return f_open(tls, _fp, _path, mode)
}

func Close(tls *libc.TLS, fp *FIL) FRESULT {
	_fp := (uintptr)(unsafe.Pointer(fp))
	return f_close(tls, _fp)
}

func Read(tls *libc.TLS, fp *FIL, buf []byte) (n int, fr FRESULT) {
	_fp := (uintptr)(unsafe.Pointer(fp))
	_buf := (uintptr)(unsafe.Pointer(&buf[0]))
	_br := (uintptr)(unsafe.Pointer(&n))
	fr = f_read(tls, _fp, _buf, uint32(len(buf)), _br)
	return n, fr
}

func Write(tls *libc.TLS, fp *FIL, buf []byte) (n int, fr FRESULT) {
	_fp := (uintptr)(unsafe.Pointer(fp))
	_buf := (uintptr)(unsafe.Pointer(&buf[0]))
	_bw := (uintptr)(unsafe.Pointer(&n))
	fr = f_write(tls, _fp, _buf, uint32(len(buf)), _bw)
	return n, fr
}

func Sync(tls *libc.TLS, fp *FIL) FRESULT {
	_fp := (uintptr)(unsafe.Pointer(fp))
	return f_sync(tls, _fp)
}

func OpenDir(tls *libc.TLS, dp *DIR, path string) FRESULT {
	_dp := (uintptr)(unsafe.Pointer(dp))
	_path, _ := libc.CString(path)
	return f_opendir(tls, _dp, _path)
}

func ReadDir(tls *libc.TLS, dp *DIR, fno *FILINFO) (fr FRESULT) {
	_dp := (uintptr)(unsafe.Pointer(dp))
	_fno := (uintptr)(unsafe.Pointer(fno))
	fr = f_readdir(tls, _dp, _fno)
	return fr
}

var RAM_disk_read = func(tls *libc.TLS, buf uintptr, sector LBA_t, count UINT) (r int32) {
	return 1
}

var MMC_disk_read = func(tls *libc.TLS, buf uintptr, sector LBA_t, count UINT) (r int32) {
	return 0
}

var USB_disk_read = func(tls *libc.TLS, buf uintptr, sector LBA_t, count UINT) (r int32) {
	return 0
}

var RAM_disk_write = func(tls *libc.TLS, buf uintptr, sector LBA_t, count UINT) (r int32) {
	return 0
}

var MMC_disk_write = func(tls *libc.TLS, buf uintptr, sector LBA_t, count UINT) (r int32) {
	return 0
}

var USB_disk_write = func(tls *libc.TLS, buf uintptr, sector LBA_t, count UINT) (r int32) {
	return 0
}

var RAM_disk_initialize = func(tls *libc.TLS) (r int32) {
	return 0
}

var MMC_disk_initialize = func(tls *libc.TLS) (r int32) {
	return 0
}

var USB_disk_initialize = func(tls *libc.TLS) (r int32) {
	return 0
}

var RAM_disk_status = func(tls *libc.TLS) (r int32) {
	return 0
}

var MMC_disk_status = func(tls *libc.TLS) (r int32) {
	return 0
}

var USB_disk_status = func(tls *libc.TLS) (r int32) {
	return 0
}

var get_fattime = func(tls *libc.TLS) (r DWORD) {
	return uint32(0)
}

/*-----------------------------------------------------------------------*/
/* Get Drive Status                                                      */
/*-----------------------------------------------------------------------*/
func disk_status(tls *libc.TLS, pdrv BYTE) (r DSTATUS) {
	var result int32
	var stat DSTATUS
	_, _ = result, stat
	switch int32(int32(pdrv)) {
	case DEV_RAM:
		result = RAM_disk_status(tls)
		// translate the reslut code here
		return stat
	case int32(DEV_MMC):
		result = MMC_disk_status(tls)
		// translate the reslut code here
		return stat
	case int32(DEV_USB):
		result = USB_disk_status(tls)
		// translate the reslut code here
		return stat
	}
	return uint8(STA_NOINIT)
}

/*-----------------------------------------------------------------------*/
/* Inidialize a Drive                                                    */
/*-----------------------------------------------------------------------*/
func disk_initialize(tls *libc.TLS, pdrv BYTE) (r DSTATUS) {
	var result int32
	var stat DSTATUS
	_, _ = result, stat
	switch int32(int32(pdrv)) {
	case DEV_RAM:
		result = RAM_disk_initialize(tls)
		// translate the reslut code here
		return stat
	case int32(DEV_MMC):
		result = MMC_disk_initialize(tls)
		// translate the reslut code here
		return stat
	case int32(DEV_USB):
		result = USB_disk_initialize(tls)
		// translate the reslut code here
		return stat
	}
	return uint8(STA_NOINIT)
}

/*-----------------------------------------------------------------------*/
/* Read Sector(s)                                                        */
/*-----------------------------------------------------------------------*/
func disk_read(tls *libc.TLS, pdrv BYTE, buff uintptr, sector LBA_t, count UINT) (r DRESULT) {
	var res DRESULT
	var result int32
	_, _ = res, result
	switch int32(int32(pdrv)) {
	case DEV_RAM:
		result = RAM_disk_read(tls, buff, sector, count)
		// translate the reslut code here
		return res
	case int32(DEV_MMC):
		result = MMC_disk_read(tls, buff, sector, count)
		// translate the reslut code here
		return res
	case int32(DEV_USB):
		result = USB_disk_read(tls, buff, sector, count)
		// translate the reslut code here
		return res
	}
	return RES_PARERR
}

/*-----------------------------------------------------------------------*/
/* Write Sector(s)                                                       */
/*-----------------------------------------------------------------------*/
func disk_write(tls *libc.TLS, pdrv BYTE, buff uintptr, sector LBA_t, count UINT) (r DRESULT) {
	var res DRESULT
	var result int32
	_, _ = res, result
	switch int32(int32(pdrv)) {
	case DEV_RAM:
		result = RAM_disk_write(tls, buff, sector, count)
		// translate the reslut code here
		return res
	case int32(DEV_MMC):
		result = MMC_disk_write(tls, buff, sector, count)
		// translate the reslut code here
		return res
	case int32(DEV_USB):
		result = USB_disk_write(tls, buff, sector, count)
		// translate the reslut code here
		return res
	}
	return RES_PARERR
}

/*-----------------------------------------------------------------------*/
/* Miscellaneous Functions                                               */
/*-----------------------------------------------------------------------*/
func disk_ioctl(tls *libc.TLS, pdrv BYTE, cmd BYTE, buff uintptr) (r DRESULT) {
	var res DRESULT
	var result int32
	_, _ = res, result
	switch int32(int32(pdrv)) {
	case DEV_RAM:
		// Process of the command for the RAM drive
		return res
	case int32(DEV_MMC):
		// Process of the command for the MMC/SD card
		return res
	case int32(DEV_USB):
		// Process of the command the USB drive
		return res
	}
	return RES_PARERR
}
