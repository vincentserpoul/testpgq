package configuration

type Config struct {
	Application struct {
		Port      int
		PrettyLog bool
		URL       struct {
			Host    string
			Schemes []string
		}
	}
	Database
	Observability struct {
		Collector struct {
			Host string
			Port int
		}
	}
	Worker struct {
		Parallel     int
		ProcessCount int
	}
}

type Database struct {
	Host         string
	Port         int
	Username     string
	Password     string
	DatabaseName string
	SSLMode      string
}
