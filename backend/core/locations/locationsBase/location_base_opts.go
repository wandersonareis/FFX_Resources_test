package locationsBase

import (
	"ffxresources/backend/common"
	"path/filepath"
)

type LocationBaseOptions struct {
	TargetDirectoryName string
	TargetDirectory     string
	GameVersionDir	  string
}

type LocationBaseOption func(*LocationBaseOptions)

func WithDirectoryName(name string) LocationBaseOption {
	return func(opts *LocationBaseOptions) {
		opts.TargetDirectoryName = name
	}
}

func WithTargetDirectory(directory string) LocationBaseOption {
	return func(opts *LocationBaseOptions) {
		opts.TargetDirectory = directory
	}
}

func WithGameVersionDir(gameVersionDir string) LocationBaseOption {
	return func(opts *LocationBaseOptions) {
		opts.GameVersionDir = gameVersionDir
	}
}

func ProcessOpts(opts []LocationBaseOption) *LocationBaseOptions {
	options := &LocationBaseOptions{}

	for _, opt := range opts {
		opt(options)
	}

	if options.TargetDirectoryName == "" {
		panic("TargetDirectoryName is required")
	}

	if options.TargetDirectory == "" {
		options.TargetDirectory = filepath.Join(common.GetExecDir(), options.TargetDirectoryName)
	}

	if options.GameVersionDir == "" {
		options.GameVersionDir = "FFX"
	}

	return options
}
