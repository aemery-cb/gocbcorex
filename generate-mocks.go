//go:generate moq -out mock_configmanager_test.go . ConfigManager
//go:generate moq -out mock_collectionresolver_test.go . CollectionResolver
//go:generate moq -out mock_vbucketrouter_test.go . VbucketRouter
//go:generate moq -out mock_kvclient_test.go . KvClient
//go:generate moq -out mock_retrymanager_test.go . RetryManager RetryController

package core
