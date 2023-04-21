/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package object

import (
	"context"
	"io"
	"time"

	libsiz "github.com/nabbar/golib/size"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type WalkFunc func(err liberr.Error, obj sdktps.Object) liberr.Error
type VersionWalkFunc func(err liberr.Error, obj sdktps.ObjectVersion) liberr.Error
type DelMakWalkFunc func(err liberr.Error, del sdktps.DeleteMarkerEntry) liberr.Error

type Object interface {
	Find(regex string) ([]string, liberr.Error)
	Size(object string) (size int64, err liberr.Error)

	List(continuationToken string) ([]sdktps.Object, string, int64, liberr.Error)
	Walk(f WalkFunc) liberr.Error

	ListPrefix(continuationToken string, prefix string) ([]sdktps.Object, string, int64, liberr.Error)
	WalkPrefix(prefix string, f WalkFunc) liberr.Error

	Head(object string) (*sdksss.HeadObjectOutput, liberr.Error)
	Get(object string) (*sdksss.GetObjectOutput, liberr.Error)
	Put(object string, body io.Reader) liberr.Error
	Delete(check bool, object string) liberr.Error
	DeleteAll(objects *sdktps.Delete) ([]sdktps.DeletedObject, liberr.Error)
	GetAttributes(object, version string) (*sdksss.GetObjectAttributesOutput, liberr.Error)

	MultipartList(keyMarker, markerId string) (uploads []sdktps.MultipartUpload, nextKeyMarker string, nextIdMarker string, count int64, e liberr.Error)
	MultipartPut(object string, body io.Reader) liberr.Error
	MultipartPutCustom(partSize libsiz.Size, object string, body io.Reader) liberr.Error
	MultipartCancel(uploadId, key string) liberr.Error

	UpdateMetadata(meta *sdksss.CopyObjectInput) liberr.Error
	SetWebsite(object, redirect string) liberr.Error

	VersionList(prefix, keyMarker, markerId string) (version []sdktps.ObjectVersion, delMarker []sdktps.DeleteMarkerEntry, nextKeyMarker, nextMarkerId string, count int64, err liberr.Error)
	VersionWalk(fv VersionWalkFunc, fd DelMakWalkFunc) liberr.Error
	VersionWalkPrefix(prefix string, fv VersionWalkFunc, fd DelMakWalkFunc) liberr.Error

	VersionGet(object, version string) (*sdksss.GetObjectOutput, liberr.Error)
	VersionHead(object, version string) (*sdksss.HeadObjectOutput, liberr.Error)
	VersionSize(object, version string) (size int64, err liberr.Error)
	VersionDelete(check bool, object, version string) liberr.Error

	GetRetention(object, version string) (until time.Time, mode string, err liberr.Error)
	SetRetention(object, version string, bypass bool, until time.Time, mode string) liberr.Error
	GetLegalHold(object, version string) (sdktps.ObjectLockLegalHoldStatus, liberr.Error)
	SetLegalHold(object, version string, flag sdktps.ObjectLockLegalHoldStatus) liberr.Error

	GetTags(object, version string) ([]sdktps.Tag, liberr.Error)
	SetTags(object, version string, tags ...sdktps.Tag) liberr.Error
}

func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Object {
	return &client{
		Helper: libhlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
