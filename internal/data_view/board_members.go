package data_view

import (
	"lsd3/internal/content"
)

type BoardMemberPageData struct {
	PageData
	Subtitle string
	Members  []content.BoardMember
}

type BoardMemberDetailData struct {
	PageData
	Member *content.BoardMember
}
