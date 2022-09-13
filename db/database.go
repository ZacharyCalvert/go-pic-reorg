package db

type trackedMedia map[string]MediaRecord

type Database struct {
	Media       trackedMedia `yaml:"media"`
	LastUpdated int64        `yaml:"updated"`
}
