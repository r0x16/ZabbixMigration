package domain

/**
 * Application provider interface
 * Represents the application itself independent of the framework
 * provides modules and bootstraps the application
 */
type ApplicationProvider interface {
	Boot()
	ProvideModules() []ApplicationModule
	Run() error
}
