package domain

// MyPostsFilter 我的发布列表筛选
type MyPostsFilter struct {
	UserID   int64
	Status   string // published / pending / rejected / deleted，空=不含已删除
	Keyword  string
	Page     int
	PageSize int
}

// MyPostStatusToDB 将前端 status 字符串映射为库内 status 值
func MyPostStatusToDB(status string) (dbStatus int, ok bool, err error) {
	switch status {
	case "published":
		return 1, true, nil
	case "pending":
		return 0, true, nil
	case "rejected":
		return 2, true, nil
	case "deleted":
		return 3, true, nil
	case "":
		return 0, false, nil
	default:
		return 0, false, ErrInvalidPostStatus
	}
}

// MyPostStatusLabel 库内 status → 前端语义
func MyPostStatusLabel(status int) string {
	switch status {
	case 1:
		return "published"
	case 0:
		return "pending"
	case 2:
		return "rejected"
	case 3:
		return "deleted"
	default:
		return "unknown"
	}
}
