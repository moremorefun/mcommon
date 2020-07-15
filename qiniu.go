package mcommon

import (
	"bytes"
	"context"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

// UploadToQiniu 上传到qiniu
func UploadToQiniu(ctx context.Context, access string, secret string, zone *storage.Zone, bucket string, fileKey string, bs []byte) error {
	buf := bytes.NewBuffer(bs)
	mac := qbox.NewMac(
		access,
		secret,
	)
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = zone
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}
	retry := 0
GotoUpload:
	err := formUploader.Put(ctx, &ret, upToken, fileKey, buf, int64(buf.Len()), &putExtra)
	if err != nil {
		qiniuErr, ok := err.(*storage.ErrorInfo)
		if ok {
			if qiniuErr.Code == 614 {
				// file exists
				return nil
			}
		}
		retry++
		if retry < 3 {
			// 重试
			goto GotoUpload
		}
		return err
	}
	return nil
}
