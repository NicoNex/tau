package tauerr

// #include "bookmark.h"
import "C"

type Bookmark = C.struct_bookmark

func NewBookmark(fileCnt string, filePos, offset int) Bookmark {
	line, lineNo, relative := line(fileCnt, filePos)

	return Bookmark{
		offset: C.int32_t(offset),
		line:   C.CString(line),
		lineno: C.int32_t(lineNo),
		pos:    C.int32_t(relative),
	}
}

func NewRawBookmark(line string, offset, lineNo, pos int) Bookmark {
	return Bookmark{
		offset: C.int32_t(offset),
		lineno: C.int32_t(lineNo),
		pos:    C.int32_t(pos),
		line:   C.CString(line),
	}
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
