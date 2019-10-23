package internal

// Config for cli
type Config struct {
	DomainNames  []string
	Path         string
	Staging      bool
	Email        string
	BucketName   string
	BucketRegion string
	Hook         string
}
