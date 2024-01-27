package fatfs

import (
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"modernc.org/libc"
)

func TestCurrent(t *testing.T) {
	runtime.LockOSThread()
	tls := libc.NewTLS()
	defer tls.Close()
	libc.SetEnviron(tls, os.Environ())
	loadVFS()
	var tests = []func(*testing.T, *libc.TLS){
		testWriteNew,
		// testReadDir,
		// testRead,
	}
	for _, test := range tests {
		test(t, tls)
	}
}

func testWriteNew(t *testing.T, tls *libc.TLS) {
	defer resetVFS()
	fss := new(FATFS)
	fr := Mount(tls, fss, "ram", 1)
	mustBeOK(t, fr)

	const mode = FA_WRITE | FA_CREATE_NEW
	var root FIL
	fr = Open(tls, &root, "deaconblues", mode)
	mustBeOK(t, fr)
}

func testReadDir(t *testing.T, tls *libc.TLS) {
	defer resetVFS()

	fss := new(FATFS)
	fr := Mount(tls, fss, "ram", 1)
	mustBeOK(t, fr)

	var dp DIR
	fr = OpenDir(tls, &dp, "rootdir")
	mustBeOK(t, fr)

	var finfo FILINFO
	fr = ReadDir(tls, &dp, &finfo)
	mustBeOK(t, fr)
}

func testRead(t *testing.T, tls *libc.TLS) {
	defer resetVFS()
	var fs FATFS
	fr := Mount(tls, &fs, "ram", 1)
	mustBeOK(t, fr)

	const mode = FA_READ
	var root FIL
	fr = Open(tls, &root, "rootfile", mode)
	mustBeOK(t, fr)

	buf := make([]byte, 512)
	n, fr := Read(tls, &root, buf)
	mustBeOK(t, fr)
	got := string(buf[:n])
	if got != rootFileContents {
		t.Errorf("rootfile contents differ got!=want\n%q\n%q\n", got, rootFileContents)
	}

	var dirfile FIL
	fr = Open(tls, &dirfile, "rootdir/dirfile", mode)
	mustBeOK(t, fr)

	n, fr = Read(tls, &dirfile, buf)
	mustBeOK(t, fr)
	got = string(buf[:n])
	if got != dirFileContents {
		t.Errorf("dirfile contents differ got!=want\n%q\n%q\n", got, dirFileContents)
	}
}

func mustBeOK(t *testing.T, fr FRESULT) {
	t.Helper()
	if fr != FR_OK {
		t.Fatal("fatalfr:", fr)
	}
}

func vfsDiff() string {
	var diff string
	for k, v := range fatInit {
		cp := fatInitCopy[k]
		if v != cp {
			diff += fmt.Sprintf("block %d differs new!=old\n%s\n%s\n", k, hex.Dump(v[:]), hex.Dump(cp[:]))
		}
	}
	return diff
}

func resetVFS() {
	fatInit = maps.Clone(fatInitCopy)
}

func loadVFS() {
	runtime.LockOSThread()
	const vfsLim = 32000 * 512
	RAM_disk_read = func(tls *libc.TLS, buf uintptr, sector, count UINT) (r int32) {
		for i := UINT(0); i < count; i++ {
			off := uintptr(i) * 512
			if off > vfsLim {
				return 1
			}
			buff := (*[512]byte)(unsafe.Pointer(buf + off))
			sec := fatInit[int64(sector)+int64(i)]
			copy(buff[:], sec[:])
		}
		return 0
	}
	RAM_disk_write = func(tls *libc.TLS, buf uintptr, sector, count UINT) (r int32) {
		for i := UINT(0); i < count; i++ {
			off := uintptr(i) * 512
			if off > vfsLim {
				return 1
			}
			buff := (*[512]byte)(unsafe.Pointer(buf + off))
			fatInit[int64(sector)+int64(i)] = *buff
		}
		return 0
	}
}

var fatInitCopy = maps.Clone(fatInit)

// Start of clean slate FAT32 filesystem image with name `keylargo`, 8GB in size.
// Contains a folder structure with a rootfile with some test, a rootdir directory
// with a file in it.
var fatInit = map[int64][512]byte{
	// Boot sector.
	0: {0xeb, 0x58, 0x90, 0x6d, 0x6b, 0x66, 0x73, 0x2e, 0x66, 0x61, 0x74, 0x00, 0x02, 0x08, 0x20, 0x00, // |.X.mkfs.fat... .|
		0x02, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x00, 0x00, 0x3e, 0x00, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, // |........>.......|
		0xd0, 0x07, 0xf0, 0x00, 0xe8, 0x3b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, // |.....;..........|
		0x01, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // |................|
		0x80, 0x00, 0x29, 0x06, 0xf1, 0x12, 0xc5, 0x6b, 0x65, 0x79, 0x6c, 0x61, 0x72, 0x67, 0x6f, 0x20, // |..)....keylargo |
		0x20, 0x20, 0x46, 0x41, 0x54, 0x33, 0x32, 0x20, 0x20, 0x20, 0x0e, 0x1f, 0xbe, 0x77, 0x7c, 0xac, // |  FAT32   ...w|.|
		0x22, 0xc0, 0x74, 0x0b, 0x56, 0xb4, 0x0e, 0xbb, 0x07, 0x00, 0xcd, 0x10, 0x5e, 0xeb, 0xf0, 0x32, // |".t.V.......^..2|
		0xe4, 0xcd, 0x16, 0xcd, 0x19, 0xeb, 0xfe, 0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x6e, // |.......This is n|
		0x6f, 0x74, 0x20, 0x61, 0x20, 0x62, 0x6f, 0x6f, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x20, 0x64, 0x69, // |ot a bootable di|
		0x73, 0x6b, 0x2e, 0x20, 0x20, 0x50, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x20, 0x69, 0x6e, 0x73, 0x65, // |sk.  Please inse|
		0x72, 0x74, 0x20, 0x61, 0x20, 0x62, 0x6f, 0x6f, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x20, 0x66, 0x6c, // |rt a bootable fl|
		0x6f, 0x70, 0x70, 0x79, 0x20, 0x61, 0x6e, 0x64, 0x0d, 0x0a, 0x70, 0x72, 0x65, 0x73, 0x73, 0x20, // |oppy and..press |
		0x61, 0x6e, 0x79, 0x20, 0x6b, 0x65, 0x79, 0x20, 0x74, 0x6f, 0x20, 0x74, 0x72, 0x79, 0x20, 0x61, // |any key to try a|
		0x67, 0x61, 0x69, 0x6e, 0x20, 0x2e, 0x2e, 0x2e, 0x20, 0x0d, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, // |gain ... .......|
		510: 0x55, 0xaa,
	},

	1: {0: 0x52, 0x52, 0x61, 0x41,
		484: 0x72, 0x72, 0x41, 0x61, 0xf8, 0xf1, 0x1d, 0x00, 0x05,
		510: 0x55, 0xaa},

	6: {0xeb, 0x58, 0x90, 0x6d, 0x6b, 0x66, 0x73, 0x2e, 0x66, 0x61, 0x74, 0x00, 0x02, 0x08, 0x20, 0x00, // |.X.mkfs.fat... .|
		0x02, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x00, 0x00, 0x3e, 0x00, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, // |........>.......|
		0xd0, 0x07, 0xf0, 0x00, 0xe8, 0x3b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, // |.....;..........|
		0x01, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // |................|
		0x80, 0x00, 0x29, 0x06, 0xf1, 0x12, 0xc5, 0x6b, 0x65, 0x79, 0x6c, 0x61, 0x72, 0x67, 0x6f, 0x20, // |..)....keylargo |
		0x20, 0x20, 0x46, 0x41, 0x54, 0x33, 0x32, 0x20, 0x20, 0x20, 0x0e, 0x1f, 0xbe, 0x77, 0x7c, 0xac, // |  FAT32   ...w|.|
		0x22, 0xc0, 0x74, 0x0b, 0x56, 0xb4, 0x0e, 0xbb, 0x07, 0x00, 0xcd, 0x10, 0x5e, 0xeb, 0xf0, 0x32, // |".t.V.......^..2|
		0xe4, 0xcd, 0x16, 0xcd, 0x19, 0xeb, 0xfe, 0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x6e, // |.......This is n|
		0x6f, 0x74, 0x20, 0x61, 0x20, 0x62, 0x6f, 0x6f, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x20, 0x64, 0x69, // |ot a bootable di|
		0x73, 0x6b, 0x2e, 0x20, 0x20, 0x50, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x20, 0x69, 0x6e, 0x73, 0x65, // |sk.  Please inse|
		0x72, 0x74, 0x20, 0x61, 0x20, 0x62, 0x6f, 0x6f, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x20, 0x66, 0x6c, // |rt a bootable fl|
		0x6f, 0x70, 0x70, 0x79, 0x20, 0x61, 0x6e, 0x64, 0x0d, 0x0a, 0x70, 0x72, 0x65, 0x73, 0x73, 0x20, // |oppy and..press |
		0x61, 0x6e, 0x79, 0x20, 0x6b, 0x65, 0x79, 0x20, 0x74, 0x6f, 0x20, 0x74, 0x72, 0x79, 0x20, 0x61, // |any key to try a|
		0x67, 0x61, 0x69, 0x6e, 0x20, 0x2e, 0x2e, 0x2e, 0x20, 0x0d, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, // |gain ... .......|
		510: 0x55, 0xaa,
	},

	7: {0: 0x52, 0x52, 0x61, 0x41,
		484: 0x72, 0x72, 0x41, 0x61, 0xfb, 0xf1, 0x1d, 0x00, 0x02,
		510: 0x55, 0xaa},

	32: {0xf8, 0xff, 0xff, 0x0f, 0xff, 0xff, 0xff, 0x0f, 0xf8, 0xff, 0xff, 0x0f, 0xff, 0xff, 0xff, 0x0f,
		0xff, 0xff, 0xff, 0x0f, 0xff, 0xff, 0xff, 0x0f},
	15368: {0xf8, 0xff, 0xff, 0x0f, 0xff, 0xff, 0xff, 0x0f, 0xf8, 0xff, 0xff, 0x0f, 0xff, 0xff, 0xff, 0x0f,
		0xff, 0xff, 0xff, 0x0f, 0xff, 0xff, 0xff, 0x0f},

	30704: { // Root directory contents.
		0x6b, 0x65, 0x79, 0x6c, 0x61, 0x72, 0x67, 0x6f, 0x20, 0x20, 0x20, 0x08, 0x00, 0x00, 0xba, 0x53, // |keylargo   ....S|
		0x35, 0x58, 0x35, 0x58, 0x00, 0x00, 0xba, 0x53, 0x35, 0x58, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // |5X5X...S5X......|
		0x41, 0x72, 0x00, 0x6f, 0x00, 0x6f, 0x00, 0x74, 0x00, 0x66, 0x00, 0x0f, 0x00, 0x1a, 0x69, 0x00, // |Ar.o.o.t.f....i.|
		0x6c, 0x00, 0x65, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, // |l.e.............|
		0x52, 0x4f, 0x4f, 0x54, 0x46, 0x49, 0x4c, 0x45, 0x20, 0x20, 0x20, 0x20, 0x00, 0x03, 0xd4, 0xbb, // |ROOTFILE    ....|
		0x37, 0x58, 0x37, 0x58, 0x00, 0x00, 0xd4, 0xbb, 0x37, 0x58, 0x04, 0x00, 0x16, 0x00, 0x00, 0x00, // |7X7X....7X......|
		0x41, 0x72, 0x00, 0x6f, 0x00, 0x6f, 0x00, 0x74, 0x00, 0x64, 0x00, 0x0f, 0x00, 0xde, 0x69, 0x00, // |Ar.o.o.t.d....i.|
		0x72, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, // |r...............|
		0x52, 0x4f, 0x4f, 0x54, 0x44, 0x49, 0x52, 0x20, 0x20, 0x20, 0x20, 0x10, 0x00, 0x29, 0xe4, 0xbb, // |ROOTDIR    ..)..|
		0x37, 0x58, 0x37, 0x58, 0x00, 0x00, 0xe4, 0xbb, 0x37, 0x58, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, // |7X7X....7X......|
		0xe5, 0x6d, 0x00, 0x2d, 0x00, 0x4e, 0x00, 0x44, 0x00, 0x38, 0x00, 0x0f, 0x00, 0x95, 0x4a, 0x00, // |.m.-.N.D.8....J.|
		0x49, 0x00, 0x32, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, // |I.2.............|
		0xe5, 0x2e, 0x00, 0x67, 0x00, 0x6f, 0x00, 0x75, 0x00, 0x74, 0x00, 0x0f, 0x00, 0x95, 0x70, 0x00, // |...g.o.u.t....p.|
		0x75, 0x00, 0x74, 0x00, 0x73, 0x00, 0x74, 0x00, 0x72, 0x00, 0x00, 0x00, 0x65, 0x00, 0x61, 0x00, // |u.t.s.t.r...e.a.|
		0xe5, 0x4f, 0x55, 0x54, 0x50, 0x55, 0x7e, 0x31, 0x20, 0x20, 0x20, 0x20, 0x00, 0x03, 0xd4, 0xbb, // |.OUTPU~1    ....|
		0x37, 0x58, 0x37, 0x58, 0x00, 0x00, 0xd4, 0xbb, 0x37, 0x58, 0x04, 0x00, 0x16, 0x00, 0x00, 0x00, // |7X7X....7X......|
	},

	30712: { // Root subdirectory "rootdir" contents.
		0x2e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x10, 0x00, 0x28, 0x64, 0xb6, // |.          ..(d.|
		0x37, 0x58, 0x37, 0x58, 0x00, 0x00, 0x64, 0xb6, 0x37, 0x58, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, // |7X7X..d.7X......|
		0x2e, 0x2e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x10, 0x00, 0x28, 0x64, 0xb6, // |..         ..(d.|
		0x37, 0x58, 0x37, 0x58, 0x00, 0x00, 0x64, 0xb6, 0x37, 0x58, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // |7X7X..d.7X......|
		0x41, 0x64, 0x00, 0x69, 0x00, 0x72, 0x00, 0x66, 0x00, 0x69, 0x00, 0x0f, 0x00, 0x27, 0x6c, 0x00, // |Ad.i.r.f.i...'l.|
		0x65, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, // |e...............|
		0x44, 0x49, 0x52, 0x46, 0x49, 0x4c, 0x45, 0x20, 0x20, 0x20, 0x20, 0x20, 0x00, 0x28, 0xe4, 0xbb, // |DIRFILE     .(..|
		0x37, 0x58, 0x37, 0x58, 0x00, 0x00, 0xe4, 0xbb, 0x37, 0x58, 0x05, 0x00, 0x49, 0x00, 0x00, 0x00, // |7X7X....7X..I...|
		0xe5, 0x6d, 0x00, 0x2d, 0x00, 0x48, 0x00, 0x49, 0x00, 0x47, 0x00, 0x0f, 0x00, 0x95, 0x37, 0x00, // |.m.-.H.I.G....7.|
		0x48, 0x00, 0x32, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, // |H.2.............|
		0xe5, 0x2e, 0x00, 0x67, 0x00, 0x6f, 0x00, 0x75, 0x00, 0x74, 0x00, 0x0f, 0x00, 0x95, 0x70, 0x00, // |...g.o.u.t....p.|
		0x75, 0x00, 0x74, 0x00, 0x73, 0x00, 0x74, 0x00, 0x72, 0x00, 0x00, 0x00, 0x65, 0x00, 0x61, 0x00, // |u.t.s.t.r...e.a.|
		0xe5, 0x4f, 0x55, 0x54, 0x50, 0x55, 0x7e, 0x31, 0x20, 0x20, 0x20, 0x20, 0x00, 0x28, 0xe4, 0xbb, // |.OUTPU~1    .(..|
		0x37, 0x58, 0x37, 0x58, 0x00, 0x00, 0xe4, 0xbb, 0x37, 0x58, 0x05, 0x00, 0x49, 0x00, 0x00, 0x00, // |7X7X....7X..I...|
	},

	// Below are file contents.

	// Says: "This is\nthe rootfile"
	30720: {0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x0a, 0x74, 0x68, 0x65, 0x20, 0x72, 0x6f, 0x6f, 0x74, 0x20, 0x66, 0x69, 0x6c, 0x65, 0x0a},
	30728: {
		0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x0a, 0x6e, 0x6f, 0x74, 0x20, // |this is not.not |
		0x74, 0x68, 0x65, 0x20, 0x72, 0x6f, 0x6f, 0x74, 0x0a, 0x6e, 0x6f, 0x74, 0x20, 0x74, 0x68, 0x65, // |the root.not the|
		0x20, 0x72, 0x6f, 0x6f, 0x74, 0x20, 0x66, 0x69, 0x6c, 0x65, 0x0a, 0x6e, 0x6f, 0x70, 0x65, 0x2e, // | root file.nope.|
		0x20, 0x0a, 0x54, 0x68, 0x69, 0x73, 0x20, 0x66, 0x69, 0x6c, 0x65, 0x20, 0x68, 0x61, 0x73, 0x20, // | .This file has |
		0x35, 0x20, 0x6c, 0x69, 0x6e, 0x65, 0x73, 0x2e, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // |5 lines.........|
	},
}

const rootFileContents = "this is\nthe root file\n"
const dirFileContents = "this is not\nnot the root\nnot the root file\nnope. \nThis file has 5 lines.\n"
