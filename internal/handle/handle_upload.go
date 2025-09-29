package handle

import (
	"errors"
	"fmt"
	"gin-blog-server/internal/global"
	"gin-blog-server/internal/model"
	"gin-blog-server/internal/utils/upload"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Upload struct{}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Success 0 {object} Response[string]
// @Router /upload/file [post]
func (*Upload) UploadFile(c *gin.Context) {
	// 获取文件头信息
	_, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		ReturnError(c, global.ErrRequest, err)
		return
	}
	// 获取 oss 对象存储接口
	oss := upload.NewOSS()
	// 实现图片上传
	filePath, _, err := oss.UploadFile(fileHeader)
	if err != nil {
		ReturnError(c, global.ErrRequest, err)
		return
	}
	ReturnSuccess(c, filePath)
}

func (*Upload) DeleteFile(c *gin.Context) {

}

func (*Upload) DownloadFile(c *gin.Context) {
	db := GetDB(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ReturnError(c, global.ErrRequest, err)
		return
	}
	// 获取文件路径
	filepath, err := model.GetFilePathById(db, id)
	if err != nil {
		ReturnError(c, global.ErrDbOp, err)
		return
	}
	// 获取文件名
	filename := filepath[strings.LastIndex(filepath, "/")+1:]
	conf := global.Conf.Upload
	// 获取文件存储位置
	storePath := conf.StorePath + "/" + filename

	slog.Debug("storePath:", slog.String("storePath", storePath))
	//打开文件
	file, err := os.Open(storePath)
	if err != nil {
		ReturnError(c, global.ErrFileOpen, err)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		ReturnError(c, global.ErrFileInfo, err)
		return
	}

	// 设置支持断点续传的响应头
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")

	//处理range请求
	rangeHeader := c.GetHeader("Range")
	if rangeHeader == "" {
		// 没有 Range 请求，返回整个文件
		c.Header("Content-Length", strconv.FormatInt(stat.Size(), 10))
		//todo: 返回文件
		ReturnSuccess(c, file)
	}

	//解析Range头
	ranges, err := parseRange(rangeHeader, stat.Size())
	if err != nil || len(ranges) == 0 {
		c.Header("Content_Range", fmt.Sprintf("bytes */%d", stat.Size()))
		ReturnError(c, global.ErrParseRange, err)
		return
	}

	// 处理 Range[0] 请求
	ra := ranges[0]
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.end, stat.Size()))
	c.Header("Content-Length", strconv.FormatInt(ra.length, 10))
	c.Status(http.StatusPartialContent)

	//读取并发送指定范围的数据
	file.Seek(ra.start, io.SeekStart)
	io.CopyN(c.Writer, file, ra.length)
}

// rangeSpecifier 表示一个范围
type rangeSpecifier struct {
	start, end, length int64
}

// parseRange 解析 Range 头
func parseRange(s string, size int64) ([]rangeSpecifier, error) {
	//格式验证
	if !strings.HasPrefix(s, "bytes=") {
		return nil, errors.New("invalid range")
	}
	var ranges []rangeSpecifier
	for _, ra := range strings.Split(s[6:], ",") {
		//处理每个范围
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}

		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("invalid range")
		}

		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		slog.Debug("start: %s, end: %s", start, end)
		var r rangeSpecifier
		//处理后缀范围（如 "-500")
		if start == "" {
			if end == "" {
				return nil, errors.New("invalid range")
			}
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.end = size - 1
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i >= size {
				return nil, errors.New("invalid range")
			}
			r.start = i
			//处理前缀范围（如 "1000-"）完整范围（如 "0-1023"
			if end == "" {
				r.end = size - 1
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.end = i
			}
		}
		r.length = r.end - r.start + 1
		ranges = append(ranges, r)
	}
	return ranges, nil
}
