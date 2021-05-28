package pipeline

import "fmt"

type Config struct {
	AppName       string
	Version       string
	CertificateID string

	WorkDir       string
	OutDir        string
	BundleDir     string
	MacOSDir      string
	FrameworksDir string
	PlugInsDir    string
	ResourcesDir  string

	Env *Environ
}

func (c *Config) initDirs(outDir, workDir string) {
	c.WorkDir = workDir
	c.OutDir = outDir
	c.BundleDir = fmt.Sprintf("%s/%s.app", outDir, c.AppName)
	c.MacOSDir = fmt.Sprintf("%s/MacOS", c.BundleDir)
	c.FrameworksDir = fmt.Sprintf("%s/Frameworks", c.BundleDir)
	c.PlugInsDir = fmt.Sprintf("%s/PlugIns", c.BundleDir)
	c.ResourcesDir = fmt.Sprintf("%s/Resources", c.BundleDir)
}

func (c *Config) initEnv() {
	c.Env = NewEnviron()

	c.Env.Setenv("OUT_DIR", c.OutDir)
	c.Env.Setenv("BUNDLE_DIR", c.BundleDir)
	c.Env.Setenv("MACOS_DIR", c.MacOSDir)
	c.Env.Setenv("BIN_DIR", c.MacOSDir)
	c.Env.Setenv("FRAMEWORKS_DIR", c.FrameworksDir)
	c.Env.Setenv("PLUGINS_DIR", c.PlugInsDir)
	c.Env.Setenv("RESOURCES_DIR", c.ResourcesDir)
}
