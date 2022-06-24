package s3minio

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"
)

var (
	TestFeature featuregate.Feature = "TestFeature"

	DefaultMutableFeatureGate featuregate.MutableFeatureGate = featuregate.NewFeatureGate()

	DefaultFeatureGate featuregate.FeatureGate = DefaultMutableFeatureGate

	defaultControllerFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
		TestFeature: {Default: true, PreRelease: featuregate.Alpha},
	}
)

func init() {
	runtime.Must(DefaultMutableFeatureGate.Add(defaultControllerFeatureGates))
}
