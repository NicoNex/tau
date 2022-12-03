package tauerr

type Bookmark struct {
	Offset int
	LineNo int
	pos    int
	Line   string
}

func NewBookmark(fileCnt string, filePos, offset int) Bookmark {
	line, lineNo, relative := line(fileCnt, filePos)

	return Bookmark{
		Offset: offset,
		Line:   line,
		LineNo: lineNo,
		pos:    relative,
	}
}
