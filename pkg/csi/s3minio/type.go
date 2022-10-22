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

	NamePrefix                                = "object.csi.aliyun.com/"
	ProvisionTypeBucketOrCreate ProvisionType = "BucketOrCreate"
	ParamProvisionTypeTag                     = NamePrefix + "provision-type"
	ParamBucketNameTag                        = NamePrefix + "bucket-name"
	ParamQuotaType                            = NamePrefix + "quota-type"
	ParamPVName                               = "csi.storage.k8s.io/pv/name"
	ParamPVCName                              = "csi.storage.k8s.io/pvc/name"
	ParamPVCNameSpace                         = "csi.storage.k8s.io/pvc/namespace"
	QuotaTypeHard                             = "hard"
	QuotaTypeFIFO                             = "fifo"
	MetaDataCapacity                          = NamePrefix + "capacity-bytes"
	MetaDataPrivisionType                     = NamePrefix + "provision-type"
	AnnoBucketName                            = NamePrefix + "bucket-name"

	SecretMinIOHost string = "host"
	SecretRegion    string = "region"
	SecretAK        string = "rootUser"
	SecretSK        string = "rootPassword"

	S3FSCmd              = "s3fs"
	S3FSPassWordFileName = ".passwd-s3fs"
	S3FSType             = "fuse.s3fs"

	maxObjectNum = 10000
)
