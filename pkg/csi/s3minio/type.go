package s3minio

// S3Config holds values to configure the driver
type S3Config struct {
	AK       string
	SK       string
	Region   string
	Endpoint string
}

type ProvisionType string

const (
	DriverName string = "s3minio"

	ProvisionFromExistBucket         ProvisionType = "ExistBucket"
	ProvisionFromDynamicCreateBucket ProvisionType = "DynamicCreateBucket"
	ParamProvisionTypeTag                          = "provisionType"
	ParamBucketNameTag                             = "bucketName"
	ParamQuotaType                                 = "quotaType"
	QuotaTypeHard                                  = "hard"
	QuotaTypeFIFO                                  = "fifo"
	MetaDataPrefix                                 = "object.csi.aliyun.com/"
	MetaDataCapacity                               = MetaDataPrefix + "capacityBytes"
	MetaDataPrivisionType                          = MetaDataPrefix + ParamProvisionTypeTag

	SecretMinIOHost string = "host"
	SecretRegion    string = "region"
	SecretAK        string = "accesskey"
	SecretSK        string = "secretkey"

	S3FSCmd              = "s3fs"
	S3FSPassWordFileName = ".passwd-s3fs"
	S3FSType             = "fuse.s3fs"

	maxObjectNum = 10000
)
