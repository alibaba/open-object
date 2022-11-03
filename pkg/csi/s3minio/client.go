package s3minio

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/tags"
	"github.com/minio/minio/pkg/madmin"
	"k8s.io/klog/v2"
)

type MinIOClient struct {
	region  string
	madmin  *madmin.AdminClient
	mclient *minio.Client
}

func NewMinIOClient(cfg *S3Config) (*MinIOClient, error) {
	// endpoint
	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	useSSL := u.Scheme == "https"
	endpoint := u.Hostname()
	if u.Port() != "" {
		endpoint = u.Hostname() + ":" + u.Port()
	}

	// minio admin
	minioAdmin, err := madmin.NewWithOptions(endpoint, &madmin.Options{
		Creds:  credentials.NewStaticV4(cfg.AK, cfg.SK, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// minio client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AK, cfg.SK, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOClient{
		madmin:  minioAdmin,
		mclient: minioClient,
		region:  cfg.Region,
	}, nil
}

func (driver *MinIOClient) CreateBucket(bucketName string, capacityBytes int64) error {
	ctx := context.Background()
	exists, err := driver.mclient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket %s exists: %s", bucketName, err.Error())
	}
	if !exists {
		if err = driver.mclient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: driver.region}); err != nil {
			return fmt.Errorf("failed to create bucket %s: %s", bucketName, err.Error())
		}
	}

	// set bucket quota
	if DefaultFeatureGate.Enabled(Quota) {
		if err = driver.SetBucketQuota(bucketName, capacityBytes, madmin.HardQuota); err != nil {
			// 创建 bucket 时设置 quota 若失败，回滚
			if err := driver.DeleteBucket(bucketName); err != nil {
				return fmt.Errorf("fail to delete bucket %s: %s", bucketName, err.Error())
			}
			return err
		}
	}

	// set bucket metadata
	bucketMap := map[string]string{
		MetaDataCapacity:      strconv.FormatInt(capacityBytes, 10),
		MetaDataPrivisionType: string(ProvisionTypeBucketOrCreate),
	}
	if err = driver.SetBucketMetadata(bucketName, bucketMap); err != nil {
		// 创建 bucket 时打 tag 若失败，回滚
		if err := driver.DeleteBucket(bucketName); err != nil {
			return fmt.Errorf("fail to delete bucket %s: %s", bucketName, err.Error())
		}
		return fmt.Errorf("failed to set bucket %s metadata: %v", bucketName, err)
	}

	return nil
}

func (driver *MinIOClient) SetBucketQuota(bucketName string, capacityBytes int64, qType madmin.QuotaType) error {
	ctx := context.Background()
	// set bucket quota restriction
	klog.Infof("quota type is %s", ParamQuotaType)
	if e := driver.madmin.SetBucketQuota(ctx, bucketName, &madmin.BucketQuota{
		Quota: uint64(capacityBytes),
		Type:  qType}); e != nil {
		return fmt.Errorf("fail to set bucket quota: %s", e.Error())
	}
	return nil
}

func (driver *MinIOClient) DeleteBucket(bucketName string) error {
	ctx := context.Background()
	exists, err := driver.mclient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		klog.Infof("bucket %s does not exist, ignoring", bucketName)
		return nil
	}

	if err := driver.EmptyBucket(bucketName); err != nil {
		return err
	}
	return driver.mclient.RemoveBucket(context.Background(), bucketName)
}

func (driver *MinIOClient) EmptyBucket(bucketName string) error {
	ctx := context.Background()
	objectsCh := make(chan minio.ObjectInfo)
	var listErr error

	go func() {
		defer close(objectsCh)

		doneCh := make(chan struct{})

		defer close(doneCh)

		// for object := range client.minio.ListObjects(bucketName, "", true, doneCh) {
		for object := range driver.mclient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			UseV1:     true,
			Recursive: true,
		}) {
			if object.Err != nil {
				listErr = object.Err
				return
			}
			objectsCh <- object
		}
	}()

	if listErr != nil {
		klog.Errorf("fail to list objects: %s", listErr.Error())
		return listErr
	}

	errorCh := driver.mclient.RemoveObjects(ctx, bucketName, objectsCh, minio.RemoveObjectsOptions{})
	for e := range errorCh {
		klog.Errorf("failed to remove object %s, error: %s", e.ObjectName, e.Err)
	}
	if len(errorCh) != 0 {
		return fmt.Errorf("failed to remove all objects of bucket %s", bucketName)
	}

	return nil
}

func (driver *MinIOClient) ListBucketObjects(bucketName string) ([]minio.ObjectInfo, error) {
	objectCh := driver.mclient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		UseV1:     true,
		Recursive: true,
	})

	var objects []minio.ObjectInfo
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// TODO: need minio version >= RELEASE.2020-05-06T23-23-25Z
// https://github.com/minio/minio/pull/9389
// store metadata in bucket tags
func (driver *MinIOClient) SetBucketMetadata(bucketName string, bucketMap map[string]string) error {
	bucketTags, err := tags.NewTags(bucketMap, false)
	if err != nil {
		return fmt.Errorf("fail to create minio bucket tags: %s", err.Error())
	}

	return driver.mclient.SetBucketTagging(context.Background(), bucketName, bucketTags)
}

func (driver *MinIOClient) GetBucketMetadata(bucketName string) (map[string]string, error) {
	bucketTags, err := driver.mclient.GetBucketTagging(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("fail to get minio bucket tags: %s", err.Error())
	}

	return bucketTags.ToMap(), nil
}

func (driver *MinIOClient) RemoveBucketMetadata(bucketName string) error {
	return driver.mclient.RemoveBucketTagging(context.Background(), bucketName)
}

func (driver *MinIOClient) GetBucketUsage(bucketName string) (int64, error) {
	objects, err := driver.ListBucketObjects(bucketName)
	if err != nil {
		return 0, err
	}
	var usage int64 = 0
	for _, object := range objects {
		usage += object.Size
	}
	return usage, nil
}

func (driver *MinIOClient) GetBucketCapacity(bucketName string) (int64, error) {
	bucketMap, err := driver.GetBucketMetadata(bucketName)
	if err != nil {
		return 0, err
	}
	capacityStr, ok := bucketMap[MetaDataCapacity]
	if !ok {
		return 0, fmt.Errorf("%s not found in bucket meta info", MetaDataCapacity)
	}
	capacity, err := strconv.ParseInt(capacityStr, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("fail to parse %s error: %s", MetaDataCapacity, err.Error())
	}
	return capacity, nil
}

func (driver *MinIOClient) FsInfo(bucketName string) (int64, int64, int64, int64, int64, int64, error) {

	var available, capacity, usage, inodes, inodesFree, inodesUsed int64
	capacity, err := driver.GetBucketCapacity(bucketName)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	usage, err = driver.GetBucketUsage(bucketName)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	available = capacity - usage
	if available < 0 {
		available = 0
	}
	inodes = maxObjectNum
	objects, err := driver.ListBucketObjects(bucketName)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	inodesUsed = int64(len(objects))
	inodesFree = inodes - inodesUsed
	if inodesFree < 0 {
		inodesFree = 0
	}

	return available, capacity, usage, inodes, inodesFree, inodesUsed, nil
}
