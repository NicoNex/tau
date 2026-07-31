package tauerr

// #include "bookmark.h"
import "C"

type Bookmark C.struct_bookmark

func NewBookmark(fileCnt string, filePos, offset int) Bookmark {
	return NewBookmarkIn("", fileCnt, filePos, offset)
}

// NewBookmarkIn is NewBookmark for an offset that came from a file other than
// the one the program is: one of the several a module is made of.
func NewBookmarkIn(file, fileCnt string, filePos, offset int) Bookmark {
	line, lineNo, relative := line(fileCnt, filePos)

	b := Bookmark{
		offset: C.int32_t(offset),
		lineno: C.int32_t(lineNo),
		pos:    C.int32_t(relative),
		len:    C.size_t(len(line)),
		line:   C.CString(line),
	}
	if file != "" {
		b.file = C.CString(file)
	}
	return b
}

func NewRawBookmark(line string, offset, lineNo, pos int) Bookmark {
	return Bookmark{
		offset: C.int32_t(offset),
		lineno: C.int32_t(lineNo),
		pos:    C.int32_t(pos),
		len:    C.size_t(len(line)),
		line:   C.CString(line),
	}
}

// File is the file the bookmark came from, empty when it is the program's own.
func (b Bookmark) File() string {
	if b.file == nil {
		return ""
	}
	return C.GoString(b.file)
}

func (b Bookmark) Offset() int {
	return int(b.offset)
}

func (b Bookmark) LineNo() int {
	return int(b.lineno)
}

func (b Bookmark) Pos() int {
	return int(b.pos)
}

func (b Bookmark) Len() int {
	return int(b.len)
}

func (b Bookmark) Line() string {
	return C.GoString(b.line)
}
