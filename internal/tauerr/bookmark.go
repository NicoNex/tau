package tauerr

type Bookmark struct {
	Offset int
	LineNo int
	Pos    int
	Line   string
}

func NewBookmark(fileCnt string, filePos, offset int) Bookmark {
	line, lineNo, relative := line(fileCnt, filePos)

	return Bookmark{
		Offset: offset,
		Line:   line,
		LineNo: lineNo,
		Pos:    relative,
	}
}
