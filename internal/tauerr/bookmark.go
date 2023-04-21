package tauerr

// #include "bookmark.h"
import "C"

type Bookmark = C.struct_bookmark

func NewBookmark(fileCnt string, filePos, offset uint) Bookmark {
	line, lineNo, relative := line(fileCnt, filePos)

	return Bookmark{
		offset: C.uint32_t(offset),
		line:   C.CString(line),
		lineno: C.uint32_t(lineNo),
		pos:    C.uint32_t(relative),
	}
}
